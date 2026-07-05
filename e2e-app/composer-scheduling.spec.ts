import { expect, test } from "@playwright/test";
import { createWorkspace, registerUser } from "./helpers";

type PostPayload = {
  workspace_id?: string;
  source_text?: string;
  content_profile?: string;
  scheduled_at?: string;
  renditions?: Array<{
    social_account_id?: string;
    profile?: string;
    body?: string;
    media?: unknown[];
  }>;
  media?: unknown[];
  [key: string]: unknown;
};

test("composer schedules a publication from the selected time", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `composer-scheduling-${unique}@example.com`;
  const postContent = "Schedule this launch note from the composer.";
  const suggestedDate = new Date(Date.now() + 48 * 60 * 60 * 1000)
    .toISOString()
    .slice(0, 10);
  const scheduledLocalTime = `${suggestedDate}T10:30`;
  let publicationPayload: PostPayload | undefined;
  let scheduleRequested = false;

  const auth = await registerUser(request, email);
  const workspaceBody = await createWorkspace(
    request,
    auth.token,
    "Composer Scheduling E2E",
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
          account_username: "openpost.bsky.social",
          account_avatar_url: "",
          instance_url: "",
          is_active: true,
          thread_replies_supported: true,
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
            connected_accounts: 1,
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
          id: "publication-schedule",
          workspace_id: publicationPayload.workspace_id,
          title: "Short text",
          content_profile: publicationPayload.content_profile,
          source_text: publicationPayload.source_text,
          status: "draft",
          scheduled_at: publicationPayload.scheduled_at ?? "",
          renditions: [],
        },
      });
      return;
    }

    await route.continue();
  });
  await page.route("**/api/v1/publications/*/schedule", async (route) => {
    if (route.request().method() === "POST") {
      scheduleRequested = true;
      await route.fulfill({
        contentType: "application/json",
        json: {
          message: "Publication scheduled",
          job_id: "job-publication-schedule",
        },
      });
      return;
    }

    await route.continue();
  });

  await page.goto("/");
  await expect(
    page.getByRole("heading", { name: "Destinations" }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: /openpost\.bsky\.social\s+Bluesky/ })
    .click();
  await page.getByLabel("Source text").fill(postContent);
  await page.locator('input[type="datetime-local"]').fill(scheduledLocalTime);
  await expect(page.getByRole("button", { name: "Schedule" })).toBeEnabled();
  await page.getByRole("button", { name: "Schedule" }).click();

  await expect(page.getByText("Publication scheduled")).toBeVisible();
  await expect.poll(() => publicationPayload).toBeTruthy();
  await expect.poll(() => scheduleRequested).toBe(true);

  expect(publicationPayload).toMatchObject({
    workspace_id: workspaceBody.id,
    content_profile: "short_text",
    source_text: postContent,
    media: [],
    renditions: [
      expect.objectContaining({
        social_account_id: "bluesky-main",
        profile: "short_text",
        body: postContent,
        media: [],
      }),
    ],
  });
  expect(publicationPayload?.scheduled_at).toBeTruthy();
  expect(new Date(publicationPayload?.scheduled_at ?? "").toString()).not.toBe(
    "Invalid Date",
  );
  expect(publicationPayload?.scheduled_at).toContain(`${suggestedDate}T`);
});
