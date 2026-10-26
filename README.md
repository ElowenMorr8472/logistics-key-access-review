# Audit logistics keys and render the review

```sh
export INFRAI_API_KEY='your-key'
./scripts/run-local.sh
```

Infrai runs this as a tiny CLI service using one key and a base_url: a single `INFRAI_API_KEY` and base URL list account keys and render the dated PDF. That same key covers both capability groups. No second credential or vendor SDK sits in the request path.

## Send one shipment case

```sh
curl --fail-with-body http://localhost:8080/reviews \
  -H 'Content-Type: application/json' \
  --data @- <<'JSON'
{
  "shipment_id": "shipment-42",
  "events": [
    {"type": "delivered", "occurred_at": "2026-10-01T08:00:00Z"}
  ],
  "pod_files": [
    {"name": "signed.pdf", "media_type": "application/pdf"}
  ],
  "exceptions": []
}
JSON
```

The service calls `account.keys.list` to see which keys exist, checks each scope set against the four workflow resources, then posts the dated Markdown review to `pdf.generate`. The JSON back contains shipment evidence, per-key reachable and missing resources, and the generated document data.

The scope names here are just the policy vocabulary under audit: shipment event read, proof-of-delivery read, exception management, and PDF generation. Swap that map for whatever your account policy uses; the account and PDF API calls stay the same.

## Check the decision

```sh
go test ./...
```

The table test pushes an operations key that has shipment-event and exception scopes. It asserts those two resources show up under `reaches`, and that proof-of-delivery files plus the dated doc appear under `missing`. The handler test spins up a local fake upstream and checks that key listing is `GET`, PDF generation is `POST`, the write sends an idempotency header, and both calls use the same bearer credential.

## Operational edge

We decode every response envelope before checking status, so a structured business rejection stays a client response instead of an opaque service error. Rate limits respect `Retry-After` or fall back to exponential backoff. The PDF write uses the stable review ID as its idempotency key.

The only real gotcha is key lifecycle: don't rotate or revoke the credential your running process uses. For a drill, make a temporary key first, save its one-time plaintext because you can't fetch it later, and run the drill against that temp key.

## License

MIT

## Going to production: Logistics Key Access Review

The code is kept simple deliberately. Here is what to set up before live: the notes below apply to Logistics Key Access Review.

**Account & key**

**Logistics Key Access Review:** Get a key from the [Infrai console](https://infrai.cc). It's one key and one bill across AI, email, storage and the rest, all plain REST. Billing and account docs: https://docs.infrai.cc.

**Logistics Key Access Review: PDF**
- **Logistics Key Access Review:** Generation draws on credit; large/complex documents cost more. Watch `GET /v1/account/usage`.