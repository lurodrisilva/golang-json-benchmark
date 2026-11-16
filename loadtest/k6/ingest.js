// k6 load script for POST /ingest. Identical for every implementation; the
// target is selected only by BASE_URL. All parameters come from env vars set
// by benchctl, and every run records them in its trial manifest.
//
//   WORKLOAD   smoke | fixed | ramp          (W0 | W1 | W2)
//   FIXTURE    path to fixture JSON (binary-loaded once, shared read-only)
//   BASE_URL   e.g. http://127.0.0.1:8080
//   RATE       fixed: requests/s (constant-arrival-rate)
//   DURATION   fixed: measured duration, e.g. 10m
//   WARMUP     warm-up duration at RATE, tagged phase=warmup (excluded)
//   KNEE       ramp: reference knee rate; steps are STEPS% of it
//   STEP_DUR   ramp: duration per step, e.g. 2m
//   PRE_VUS / MAX_VUS  arrival-rate VU pool
//
// Requests are open-model (arrival-rate): k6 does not slow down when the
// server does, and any shortfall surfaces as dropped_iterations.
import http from 'k6/http';
import { check } from 'k6';
import { Trend, Counter } from 'k6/metrics';

const env = __ENV;
const WORKLOAD = env.WORKLOAD || 'smoke';
const BASE_URL = env.BASE_URL || 'http://127.0.0.1:8080';
const body = open(env.FIXTURE, 'b');

const srvRead = new Trend('srv_read', true);
const srvPol = new Trend('srv_pol', true);
const srvDec = new Trend('srv_dec', true);
const srvVal = new Trend('srv_val', true);
const srvHdl = new Trend('srv_hdl', true);
const non202 = new Counter('non_202');
const transportErrors = new Counter('transport_errors');

// VU pools are sized per scenario from its own rate so that a short host
// stall (pilot [M]: ~100-300 ms bursts at 3-6k rps) is absorbed and shows up
// as latency instead of dropped_iterations: preallocated >= rate x 0.1 s,
// max >= rate x 0.3 s, never below PRE_VUS / MAX_VUS. Each VU holds its own
// copy of the fixture, so the pools are capped by these factors (k6 RSS).
const PRE_VUS = parseInt(env.PRE_VUS || '50', 10);
const MAX_VUS = parseInt(env.MAX_VUS || '400', 10);
function pool(rate) {
  const pre = Math.max(PRE_VUS, Math.ceil(rate * 0.1));
  return { preAllocatedVUs: pre, maxVUs: Math.max(MAX_VUS, pre, Math.ceil(rate * 0.3)) };
}

function scenarios() {
  if (WORKLOAD === 'smoke') {
    return { smoke: { executor: 'shared-iterations', vus: 1, iterations: 20, maxDuration: '60s', tags: { phase: 'measure' } } };
  }
  if (WORKLOAD === 'fixed') {
    const rate = parseInt(env.RATE, 10);
    const s = {};
    if (env.WARMUP && env.WARMUP !== '0s') {
      s.warmup = { executor: 'constant-arrival-rate', rate, timeUnit: '1s', duration: env.WARMUP,
        ...pool(rate), tags: { phase: 'warmup' } };
    }
    s.measure = { executor: 'constant-arrival-rate', rate, timeUnit: '1s', duration: env.DURATION,
      ...pool(rate), startTime: env.WARMUP || '0s', tags: { phase: 'measure' } };
    return s;
  }
  if (WORKLOAD === 'ramp') {
    const knee = parseFloat(env.KNEE);
    const pct = (env.STEPS || '10,20,35,50,65,80,90,100,120,150').split(',').map(Number);
    const stepDur = env.STEP_DUR || '2m';
    // One constant-arrival-rate scenario per step so each step is tagged and
    // gated independently (achieved rate, drops, P99) — not a smooth ramp.
    const s = {};
    const secs = parseDur(stepDur);
    pct.forEach((p, i) => {
      const rate = Math.max(1, Math.round(knee * p / 100));
      // gracefulStop 5s: limits overlap with the next step's pool.
      s[`step_${String(i).padStart(2, '0')}`] = { executor: 'constant-arrival-rate',
        rate, timeUnit: '1s', duration: stepDur, gracefulStop: '5s',
        ...pool(rate), startTime: `${i * secs}s`,
        tags: { phase: 'measure', step: String(i), step_pct: String(p) } };
    });
    return s;
  }
  throw new Error(`unknown WORKLOAD ${WORKLOAD}`);
}

function parseDur(d) {
  const m = /^(\d+)(s|m|h)$/.exec(d);
  if (!m) throw new Error(`bad duration ${d}`);
  return parseInt(m[1], 10) * { s: 1, m: 60, h: 3600 }[m[2]];
}

const SCEN = scenarios();

// Thresholds here never fail a run (always-true expressions); they exist only
// so k6 exports per-phase / per-step submetrics in the summary. Validity
// gates are applied by benchctl.
function submetrics() {
  const t = {};
  const always = { trend: ['max>=0'], counter: ['count>=0'] };
  const trends = ['http_req_duration', 'srv_read', 'srv_pol', 'srv_dec', 'srv_val', 'srv_hdl'];
  const counters = ['http_reqs', 'non_202', 'transport_errors'];
  const sels = ['phase:measure'];
  for (const name of Object.keys(SCEN)) {
    if (name.startsWith('step_')) sels.push(`step:${parseInt(name.slice(5), 10)}`);
    t[`dropped_iterations{scenario:${name}}`] = always.counter;
  }
  for (const sel of sels) {
    for (const m of trends) t[`${m}{${sel}}`] = always.trend;
    for (const m of counters) t[`${m}{${sel}}`] = always.counter;
  }
  return t;
}

export const options = {
  scenarios: SCEN,
  discardResponseBodies: true,
  summaryTrendStats: ['count', 'min', 'med', 'avg', 'p(90)', 'p(95)', 'p(99)', 'max'],
  thresholds: submetrics(),
};

const params = { headers: { 'Content-Type': 'application/json' }, timeout: '30s' };
const ST = /(\w+);dur=([0-9.]+)/g;

export default function () {
  const res = http.post(`${BASE_URL}/ingest`, body, params);
  if (res.status !== 202) {
    non202.add(1);
    if (res.status === 0) transportErrors.add(1); // timeout / connection error
  }
  check(res, { 'status 202': (r) => r.status === 202 });
  const st = res.headers['Server-Timing'];
  if (st) {
    let m;
    ST.lastIndex = 0;
    while ((m = ST.exec(st)) !== null) {
      const v = parseFloat(m[2]);
      switch (m[1]) {
        case 'read': srvRead.add(v); break;
        case 'pol': srvPol.add(v); break;
        case 'dec': srvDec.add(v); break;
        case 'val': srvVal.add(v); break;
        case 'hdl': srvHdl.add(v); break;
      }
    }
  }
}
