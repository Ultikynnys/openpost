import { expect, test } from "@playwright/test";
import { createWorkspace, registerUser } from "./helpers";

test("composer renders account-specific renditions", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `composer-preview-${unique}@example.com`;

  const auth = await registerUser(request, email);
  const workspaceBody = await createWorkspace(
    request,
    auth.token,
    "Composer Preview E2E",
  );

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.route("**/api/v1/accounts?**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        {
          id: "bluesky-main",
          slug: "bluesky-main",
          platform: "bluesky",
          account_id: "bsky-main",
          account_username: "openpost_main",
          account_avatar_url: "https://cdn.example/main.jpg",
          instance_url: "",
          is_active: true,
          thread_replies_supported: false,
        },
        {
          id: "bluesky-studio",
          slug: "bluesky-studio",
          platform: "bluesky",
          account_id: "bsky-studio",
          account_username: "openpost_studio",
          account_avatar_url: "https://cdn.example/studio.jpg",
          instance_url: "",
          is_active: true,
          thread_replies_supported: false,
        },
      ],
    });
  });
  await page.route("**/api/v1/provider-readiness?**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: {
        providers: [
          {
            provider: "bluesky",
            configured_app_state: "ready",
            connected_accounts: 2,
            blocking_issues: [],
            next_actions: [],
          },
        ],
      },
    });
  });

  await page.goto("/");
  await expect(
    page.getByRole("heading", { name: "Destinations" }),
  ).toBeVisible();
  await page.getByRole("button", { name: /openpost_main\s+Bluesky/ }).click();
  await page.getByRole("button", { name: /openpost_studio\s+Bluesky/ }).click();
  await page.getByLabel("Source text").fill("Launch update");

  await expect(page.locator('[data-testid="instagram-preview"]')).toHaveCount(
    0,
  );
  await expect(
    page.getByRole("button", { name: /openpost_main\s+Bluesky post/ }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: /openpost_studio\s+Bluesky post/ }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: /openpost_studio\s+Bluesky post/ })
    .click();
  await expect(page.locator("#rendition-body")).toHaveValue("Launch update");
});
