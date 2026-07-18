import { expect, test } from "@playwright/test";
import { authenticatePage, createWorkspace, registerUser } from "./helpers";

test("authenticated navigation keeps the app shell mounted", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `app-shell-${unique}@example.com`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "Persistent Shell E2E");
  await authenticatePage(page, auth.token);
  await page.goto("/");
  await expect(page.getByTestId("app-sidebar")).toBeVisible();

  await page.evaluate(() => {
    const shell = document.querySelector('[data-testid="app-sidebar"]');
    if (!shell) throw new Error("App sidebar was not mounted");
    (
      window as Window & { __openpostShellRemoved?: boolean }
    ).__openpostShellRemoved = false;
    new MutationObserver(() => {
      if (!shell.isConnected) {
        (
          window as Window & { __openpostShellRemoved?: boolean }
        ).__openpostShellRemoved = true;
      }
    }).observe(document.body, { childList: true, subtree: true });
  });

  await page.getByRole("button", { name: "Posts", exact: true }).click();
  await expect(page).toHaveURL(/\/activity$/);
  await expect(page.getByTestId("app-sidebar")).toBeVisible();
  await expect(page.getByTestId("sidebar-new-post")).toBeVisible();
  await expect
    .poll(() =>
      page.evaluate(
        () =>
          (window as Window & { __openpostShellRemoved?: boolean })
            .__openpostShellRemoved ?? false,
      ),
    )
    .toBe(false);

  await page.getByRole("button", { name: "Calendar", exact: true }).click();
  await expect(page).toHaveURL(/\/calendar$/);
  await expect(page.getByTestId("app-sidebar")).toBeVisible();
});

test("first autosave establishes the draft URL and keeps draft actions in one composer", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `composer-draft-${unique}@example.com`;
  const content = "Keep this draft attached to its own URL.";

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "Draft URL E2E");
  await authenticatePage(page, auth.token);
  await page.goto("/");

  await expect(page.getByTestId("sidebar-home-brand")).toBeVisible();
  await expect(page.getByTestId("sidebar-new-post")).toHaveCount(0);

  await page.getByLabel("Post text").fill(content);
  await expect(page).toHaveURL(/\/posts\/[a-zA-Z0-9-]+$/, { timeout: 10_000 });
  await expect(page.getByTestId("sidebar-new-post")).toBeVisible();
  await expect(page.getByTestId("composer-delete")).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Schedule", exact: true }).first(),
  ).toBeVisible();
  await expect(page.getByText("Editing draft post")).toHaveCount(0);
  await expect(page.getByText("Saved", { exact: true })).toHaveCount(0);

  await page.reload();
  await expect(page.getByLabel("Post text")).toHaveValue(content);
  await expect(page.getByTestId("composer-delete")).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Schedule", exact: true }).first(),
  ).toBeVisible();
});

test("collapsed sidebar keeps the OpenPost mark without overflowing text", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `sidebar-logo-${unique}@example.com`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "Sidebar Logo E2E");
  await authenticatePage(page, auth.token);
  await page.goto("/");

  await page.getByRole("button", { name: "Toggle Sidebar" }).click();
  const home = page.getByRole("link", { name: "OpenPost home" });
  await expect(home.locator("svg")).toBeVisible();
  await expect(home.getByText("OpenPost", { exact: true })).toHaveCount(0);
  await expect(page.getByTestId("desktop-sidebar-planner")).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "Posts", exact: true }),
  ).toBeVisible();
});

test("desktop planning sidebar resumes drafts and stays out of mobile navigation", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `sidebar-planner-${unique}@example.com`;
  const auth = await registerUser(request, email);
  const workspace = (await createWorkspace(
    request,
    auth.token,
    "Planning Sidebar E2E",
  )) as { id: string };
  const draft = await request.post("/api/v1/posts", {
    headers: { Authorization: `Bearer ${auth.token}` },
    data: {
      workspace_id: workspace.id,
      content: "Resume the launch announcement",
      social_account_ids: [],
      media_ids: [],
    },
  });
  expect(draft.ok()).toBeTruthy();
  const draftBody = (await draft.json()) as { id: string };

  await authenticatePage(page, auth.token);
  await page.goto("/");

  const planner = page.getByTestId("desktop-sidebar-planner");
  await expect(planner).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Calendar", exact: true }),
  ).toBeVisible();
  await expect(page.getByText("Resume the launch announcement")).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Media", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Accounts", exact: true }),
  ).toBeVisible();

  await page
    .getByRole("button", {
      name: "Resume draft: Resume the launch announcement",
    })
    .click();
  await expect(page).toHaveURL(new RegExp(`/posts/${draftBody.id}$`));

  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/");
  await expect(planner).toHaveCount(0);
  await expect(
    page.getByRole("navigation", { name: "Primary navigation" }),
  ).toBeVisible();
});
