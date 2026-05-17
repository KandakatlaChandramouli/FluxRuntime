// Phase 1 — Warm-Up
// Duration: 0s → 30s
// Traffic:  0 → 5k RPS gradual ramp
// Distribution: multi-key randomized

import grpc from 'k6/net/grpc';
import { check } from 'k6';
import { BASE_URL, WARMUP_RPS } from './config.js';

const client = new grpc.Client();
client.load(['../../proto'], 'ticket/v1/reservation.proto');

export const options = {
  scenarios: {
    warmup: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 500,
      stages: [
        { target: WARMUP_RPS, duration: '30s' },
      ],
    },
  },
  thresholds: {
    grpc_req_duration: ['p(99)<10000', 'p(95)<5000'],
  },
};

export function setup() {
  client.connect(BASE_URL, { plaintext: true });
}

export default function () {
  // Randomized multi-key distribution to warm queues and Redis uniformly.
  const eventId = Math.floor(Math.random() * 1000) + 1;

  const resp = client.invoke('ticket.v1.TicketReservationService/ReserveTicket', {
    event_id: eventId,
    quantity: 1,
    user_id_high: 0,
    user_id_low: Math.floor(Math.random() * 1000000),
  });

  check(resp, {
    'status is not unspecified': (r) =>
      r.message.status !== 'STATUS_UNSPECIFIED',
  });
}

export function teardown() {
  client.close();
}
