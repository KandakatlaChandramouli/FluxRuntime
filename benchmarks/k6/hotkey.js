// Phase 2 — Contention Shockwave
// Duration: 30s → 60s
// Traffic:  50k+ RPS spike
// Skew:     99% against one EventID
// Inventory: 100 tickets

import grpc from 'k6/net/grpc';
import { check } from 'k6';
import { BASE_URL, PEAK_RPS, TARGET_EVENT_HOT } from './config.js';

const client = new grpc.Client();
client.load(['../../proto'], 'ticket/v1/reservation.proto');

export const options = {
  scenarios: {
    contention_shockwave: {
      executor: 'constant-arrival-rate',
      rate: PEAK_RPS,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 2000,
      maxVUs: 5000,
    },
  },
  thresholds: {
    grpc_req_duration: ['p(99)<10000', 'p(95)<5000'],
    'grpc_req_duration{status:STATUS_REJECTED}': ['p(99)<1000'],
  },
};

export function setup() {
  client.connect(BASE_URL, { plaintext: true });
}

export default function () {
  // 99% hot key, 1% random — forces measurable shard imbalance.
  const hot = Math.random() < 0.99;
  const eventId = hot ? TARGET_EVENT_HOT : Math.floor(Math.random() * 1000) + 1;

  const resp = client.invoke('ticket.v1.TicketReservationService/ReserveTicket', {
    event_id: eventId,
    quantity: 1,
    user_id_high: 0,
    user_id_low: Math.floor(Math.random() * 1000000),
  });

  check(resp, {
    'no STATUS_UNSPECIFIED': (r) => r.message.status !== 'STATUS_UNSPECIFIED',
    'queue rejection under 1ms': (r) =>
      r.message.status !== 'STATUS_REJECTED' || r.timings.duration < 1,
  });
}

export function teardown() {
  client.close();
}
