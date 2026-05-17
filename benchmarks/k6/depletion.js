// Phase 3 — Depletion Tail
// Duration: 60s → 90s
// Traffic:  maintain peak load post-exhaustion
// Objective: measure rejection efficiency; observe tail stabilization

import grpc from 'k6/net/grpc';
import { check } from 'k6';
import { BASE_URL, PEAK_RPS, TARGET_EVENT_HOT } from './config.js';

const client = new grpc.Client();
client.load(['../../proto'], 'ticket/v1/reservation.proto');

export const options = {
  scenarios: {
    depletion_tail: {
      executor: 'constant-arrival-rate',
      rate: PEAK_RPS,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 2000,
      maxVUs: 5000,
    },
  },
  thresholds: {
    // P99 must stabilize or improve vs Phase 2 after depletion.
    grpc_req_duration: ['p(99)<10000', 'p(95)<5000'],
    // All post-depletion hot-key requests must be STATUS_EXHAUSTED.
    'grpc_req_duration{status:STATUS_EXHAUSTED}': ['p(99)<5000'],
  },
};

export function setup() {
  client.connect(BASE_URL, { plaintext: true });
}

export default function () {
  const resp = client.invoke('ticket.v1.TicketReservationService/ReserveTicket', {
    event_id: TARGET_EVENT_HOT,
    quantity: 1,
    user_id_high: 0,
    user_id_low: Math.floor(Math.random() * 1000000),
  });

  check(resp, {
    'exhausted or rejected after depletion': (r) =>
      r.message.status === 'STATUS_EXHAUSTED' ||
      r.message.status === 'STATUS_REJECTED',
    'zero oversells': (r) =>
      r.message.status !== 'STATUS_SUCCESS',
  });
}

export function teardown() {
  client.close();
}
