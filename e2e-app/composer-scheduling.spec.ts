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
  await page.route("**/api/v1/publications/*/validate", async (route) => {
    if (route.request().method() === "POST") {
      await route.fulfill({
        contentType: "application/json",
        json: {
          publication_id: "publication-schedule",
          issues: [],
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
  await page.getByRole("button", { name: "Link" }).click();
  await expect(page.getByRole("heading", { name: "Accounts" })).toBeVisible();
  await page.getByLabel("Link URL").fill("https://openpost.social/launch");
  await page.getByLabel("Post text").fill(postContent);
  await page.getByRole("button", { name: "Schedule" }).first().click();
  await page.getByRole("button", { name: "10:30" }).click();
  await expect(page.getByRole("button", { name: "Schedule" })).toBeEnabled();
  await page.getByRole("button", { name: "Schedule" }).click();

  await expect(page.getByText("Publication scheduled")).toBeVisible();
  await expect.poll(() => publicationPayload).toBeTruthy();
  await expect.poll(() => scheduleRequested).toBe(true);

  expect(publicationPayload).toMatchObject({
    workspace_id: workspaceBody.id,
    content_profile: "link_share",
    source_text: postContent,
    source_url: "https://openpost.social/launch",
    media: [],
    renditions: [
      expect.objectContaining({
        social_account_id: "bluesky-main",
        profile: "link_share",
        body: postContent,
        settings: expect.objectContaining({
          url: "https://openpost.social/launch",
        }),
        media: [],
      }),
    ],
  });
  expect(publicationPayload?.scheduled_at).toBeTruthy();
  expect(new Date(publicationPayload?.scheduled_at ?? "").toString()).not.toBe(
    "Invalid Date",
  );
});
