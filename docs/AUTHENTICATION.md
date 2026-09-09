# Authentication contract

`POST /cli/auth/device` accepts `device_name` and returns `device_code`, `user_code`, verification URLs, `expires_in=600`, and `interval=5`. The device code is stored only as an HMAC digest. Creation and polling are rate limited.

The signed-in website posts `{code, decision}` to `POST /management/cli/authorize`. A GET never approves a device. The CLI polls `POST /cli/auth/token` with the device grant. Pending, slow, denied, expired and used grants return OAuth-style error codes.

Approval issues a 15-minute gateway credential and a rotating refresh token. The session has a 30-day absolute lifetime. Refresh tokens are HMAC digests; reuse revokes the whole session. `POST /cli/auth/revoke` and `DELETE /management/cli/devices/:id` revoke the active gateway credential. Supabase cookies and refresh tokens never leave the website.

The CLI stores credentials with Windows DPAPI, macOS Keychain or Linux Secret Service. If secure storage is unavailable, it directs the user to a process-only `TORVE_API_KEY` rather than writing plaintext.
