import { expect, test } from "@playwright/test";
import { createWorkspace, registerUser } from "./helpers";

type PostPayload = {
  workspace_id?: string;
  source_text?: string;
  source_url?: string;
  content_profile?: string;
  renditions?: Array<{
    social_account_id?: string;
    profile?: string;
    body?: string;
    settings?: Record<string, unknown>;
  }>;
  [key: string]: unknown;
};

test("composer renders account-specific renditions", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `composer-preview-${unique}@example.com`;
  let publicationPayload: PostPayload | undefined;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "Composer Preview E2E");

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
  await page.route("**/api/v1/publications", async (route) => {
    if (route.request().method() === "POST") {
      publicationPayload = JSON.parse(
        route.request().postData() ?? "{}",
      ) as PostPayload;

      await route.fulfill({
        contentType: "application/json",
        json: {
          id: "publication-preview",
          workspace_id: publicationPayload.workspace_id,
          title: "Launch update",
          content_profile: publicationPayload.content_profile,
          source_text: publicationPayload.source_text,
          source_url: publicationPayload.source_url,
          status: "draft",
          renditions: [],
        },
      });
      return;
    }

    await route.continue();
  });

  await page.goto("/");
  await page.getByRole("button", { name: "Link" }).click();
  await expect(page.getByRole("heading", { name: "Accounts" })).toBeVisible();
  await page.getByLabel("Link URL").fill("https://openpost.social/launch");
  await page.getByLabel("Post text").fill("Launch update");

  await expect(page.locator('[data-testid="instagram-preview"]')).toHaveCount(
    0,
  );
  await expect(
    page.getByRole("button", { name: /openpost_main\s+Bluesky/ }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: /openpost_studio\s+Bluesky/ }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Save draft" }).click();
  await expect(page.getByText("Draft saved")).toBeVisible();
  await expect.poll(() => publicationPayload).toBeTruthy();

  expect(publicationPayload).toMatchObject({
    content_profile: "link_share",
    source_text: "Launch update",
    source_url: "https://openpost.social/launch",
    renditions: [
      expect.objectContaining({
        social_account_id: "bluesky-main",
        profile: "link_share",
        body: "Launch update",
        settings: expect.objectContaining({
          url: "https://openpost.social/launch",
        }),
      }),
      expect.objectContaining({
        social_account_id: "bluesky-studio",
        profile: "link_share",
        body: "Launch update",
        settings: expect.objectContaining({
          url: "https://openpost.social/launch",
        }),
      }),
    ],
  });
});
