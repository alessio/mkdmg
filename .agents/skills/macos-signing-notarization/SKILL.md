---
name: macos-signing-notarization
description: >-
  Manage Apple Developer code signing and notarization workflows for mkdmg.
  Use this skill when configuring Developer ID certificates, keychain profiles, notarytool integration, and DMG ticket stapling.
---

# macOS Code Signing & Notarization Runbook

This skill outlines how to configure, execute, and verify code signing and notarization for disk images created with `mkdmg`.

---

## 1. Prerequisites

### Code Signing Identity
To list installed Apple signing identities on the host:
```sh
security find-identity -v -p codesigning
```
Look for identities named:
`Developer ID Application: Your Name/Company (TEAMID)`

### Storing Notarization Credentials
Store your Apple ID credentials securely in macOS Keychain using `notarytool`:
```sh
xcrun notarytool store-credentials "AC_PASSWORD" \
  --apple-id "developer@example.com" \
  --team-id "TEAMID1234" \
  --password "xxxx-xxxx-xxxx-xxxx"
```

---

## 2. Configuration in `mkdmg.json`

Set `signing_identity` and `notarize_credentials` in your configuration file:

```json
{
  "output_path": "./dist/MyApp.dmg",
  "source_dir": "./build/Release",
  "volume_name": "MyApp",
  "signing_identity": "Developer ID Application: Example Corp (TEAMID1234)",
  "notarize_credentials": "AC_PASSWORD"
}
```

---

## 3. Execution Sequence in `mkdmg`

When configured, `mkdmg` performs:
1. **DMG Creation**: Creates and compresses the final `.dmg`.
2. **Codesigning**: Executes:
   ```sh
   codesign --sign "Developer ID Application: ..." --deep --strict --options runtime ./dist/MyApp.dmg
   ```
3. **Notarization Submission**: Submits to Apple Notary Service:
   ```sh
   xcrun notarytool submit ./dist/MyApp.dmg --keychain-profile "AC_PASSWORD" --wait
   ```
4. **Stapling**: Attaches the notarization ticket to the DMG:
   ```sh
   xcrun stapler staple ./dist/MyApp.dmg
   ```

---

## 4. Verification

To verify the signed and notarized DMG:
```sh
# Verify signature
codesign --verify --verbose=4 ./dist/MyApp.dmg

# Verify Gatekeeper assessment
spctl --assess --type open --context context:primary-signature --verbose ./dist/MyApp.dmg

# Verify stapler ticket
xcrun stapler validate ./dist/MyApp.dmg
```
