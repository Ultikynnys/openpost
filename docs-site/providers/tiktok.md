# TikTok

TikTok support is available as an initial video/photo publishing slice. It uses OAuth plus the Content Posting API direct-post and upload flows.

## What you need

- TikTok developer app
- Login Kit and Content Posting API access
- Provider app registry entry with provider `tiktok`
- Callback URL: `https://your-domain.com/api/v1/accounts/tiktok/callback`
- Public `OPENPOST_MEDIA_URL` or S3/R2 public media URL for Direct Post media URLs
- Scopes: `user.info.basic`, `user.info.profile`, `video.publish`, `video.upload`, and photo-post access when using image posts

Example `OPENPOST_PROVIDER_APPS` entry:

```json
[
  {
    "provider": "tiktok",
    "client_id": "your-client-key",
    "client_secret": "your-client-secret",
    "redirect_uri": "https://your-domain.com/api/v1/accounts/tiktok/callback"
  }
]
```

## Current scope and limits

- Supports one video attachment for Direct Post.
- Supports inbox/upload mode for video when configured.
- Supports image/photo posts when all attached media are images and TikTok app access allows the photo-post path.
- Text-only posts are not supported.
- Direct Post media URLs must be public HTTPS.
- Live-account and app-review behavior still needs deployment verification.

## Common issues

- `OPENPOST_MEDIA_URL` points at localhost or a private host.
- TikTok app lacks Content Posting API access or required scopes.
- The TikTok app's redirect URI does not exactly match OpenPost's callback URL.
