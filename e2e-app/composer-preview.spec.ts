import { expect, test } from "@playwright/test";
import { authenticatePage, createWorkspace, registerUser } from "./helpers";

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

  await authenticatePage(page, auth.token);
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
  await page.getByTestId("composer-mode-select").click();
  await page.getByRole("option", { name: "Short video" }).click();
  await expect(
    page.getByRole("button", { name: "Target accounts" }),
  ).toBeVisible();
  await expect(page.getByLabel("Composer workspace")).toHaveCount(0);
  await page.getByLabel("Caption").fill("Launch update");

  await expect(page.locator('[data-testid="instagram-preview"]')).toHaveCount(
    0,
  );
  await expect(page.getByLabel(/Remove .* from targets/)).toHaveCount(0);
  const accountControl = page.getByTestId("composer-account-control");
  await expect(accountControl.getByTestId("composer-account-icon")).toHaveCount(
    2,
  );
  await accountControl.click();
  await expect(page.getByTestId("composer-account-row")).toHaveCount(2);
  await expect(
    page.getByText("Bluesky @openpost_main", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("Bluesky @openpost_studio", { exact: true }),
  ).toBeVisible();
  await page.keyboard.press("Escape");
  await page.getByRole("button", { name: "Save draft" }).click();
  await expect(page.getByText("Draft saved")).toBeVisible();
  await expect.poll(() => publicationPayload).toBeTruthy();

  expect(publicationPayload).toMatchObject({
    content_profile: "short_video",
    source_text: "Launch update",
    renditions: [
      expect.objectContaining({
        social_account_id: "bluesky-main",
        profile: "short_video",
        body: "Launch update",
      }),
      expect.objectContaining({
        social_account_id: "bluesky-studio",
        profile: "short_video",
        body: "Launch update",
      }),
    ],
  });
  expect(publicationPayload?.source_url).toBeUndefined();
  for (const rendition of publicationPayload?.renditions ?? []) {
    expect(rendition.settings).not.toHaveProperty("url");
    expect(rendition.settings?.link_url ?? "").toBe("");
  }
});
