/* eslint-disable no-secrets/no-secrets -- fake JWTs for testing, synthetic header/payload, no real signature (signature is literal "mock"), decoded payload contains only test@run.ai and a fake issuer */

// Decoded payload: { email: "test@run.ai", preferred_username: "test@run.ai", sub: "00000000-...", iss: "https://fake-issuer.example.com/..." }
// Used by: cluster-new.test.ts, cluster-installation-step.test.ts
export const FAKE_ACCESS_TOKEN_A =
  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMCwiaXNzIjoiaHR0cHM6Ly9mYWtlLWlzc3Vlci5leGFtcGxlLmNvbS9hdXRoL3JlYWxtcy9ydW5haSIsImF1ZCI6WyJydW5haS1hZG1pbi11aSIsInJ1bmFpIl0sInN1YiI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInR5cCI6IklEIiwiYXpwIjoicnVuYWktYWRtaW4tdWkiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwicHJlZmVycmVkX3VzZXJuYW1lIjoidGVzdEBydW4uYWkiLCJlbWFpbCI6InRlc3RAcnVuLmFpIn0.mock";

// Decoded payload: { email: "test@run.ai", preferred_username: "test@run.ai", tenant_id: 1, iss: "https://fake-issuer.example.com/..." }
// Used by: workload-template-new.mock.ts, workload-template-create-form.test.ts
export const FAKE_ACCESS_TOKEN_B =
  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMCwiaXNzIjoiaHR0cHM6Ly9mYWtlLWlzc3Vlci5leGFtcGxlLmNvbS9hdXRoL3JlYWxtcy9ydW5haSIsImF1ZCI6WyJydW5haS1hZG1pbi11aSIsInJ1bmFpIl0sInN1YiI6IjExMTExMTExLTExMTEtMTExMS0xMTExLTExMTExMTExMTExMSIsInR5cCI6IklEIiwiYXpwIjoicnVuYWktYWRtaW4tdWkiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwicHJlZmVycmVkX3VzZXJuYW1lIjoidGVzdEBydW4uYWkiLCJlbWFpbCI6InRlc3RAcnVuLmFpIiwidGVuYW50X2lkIjoxfQ.mock";

// Fake JWT — payload: { exp: 9999999999, iat: 1600000000 }, signature is literal "mock"
// Used by: auth-axios.service.test.ts
export const FAKE_ACCESS_TOKEN_C = "eyJhbGciOiJSUzI1NiJ9.eyJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMH0.mock";

/* eslint-enable no-secrets/no-secrets */
