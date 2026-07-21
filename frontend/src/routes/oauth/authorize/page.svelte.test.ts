import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AuthorizePage from './+page.svelte';

const mocks = vi.hoisted(() => {
	type PageValue = { url: URL };
	type Subscriber<T> = (value: T) => void;

	const pageValue: PageValue = {
		url: new URL(
			'http://localhost/oauth/authorize?response_type=code&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&code_challenge=challenge&code_challenge_method=S256'
		)
	};
	const pageStore = {
		subscribe(run: Subscriber<PageValue>) {
			run(pageValue);
			return () => undefined;
		}
	};
	const authValue = {
		user: { id: 'user-1' },
		isLoading: false,
		isAuthenticated: true
	};
	const authStore = {
		subscribe(run: Subscriber<typeof authValue>) {
			run(authValue);
			return () => undefined;
		}
	};

	return {
		goto: vi.fn(),
		post: vi.fn(),
		pageStore,
		authStore,
		workspaceCtx: {
			currentWorkspace: { id: 'workspace-1', name: 'Workspace' }
		}
	};
});

vi.mock('$app/stores', () => ({ page: mocks.pageStore }));
vi.mock('$app/navigation', () => ({ goto: mocks.goto }));
vi.mock('$lib/stores/auth', () => ({ auth: mocks.authStore }));
vi.mock('$lib/stores/workspace.svelte', () => ({ workspaceCtx: mocks.workspaceCtx }));
vi.mock('$lib/api/client', () => ({ client: { POST: mocks.post } }));

describe('OAuth authorization request validation', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('disables denial when the request has no client ID', async () => {
		const screen = await render(AuthorizePage);

		await expect
			.element(screen.getByText('This OAuth request is missing a client ID.'))
			.toBeVisible();
		await expect.element(screen.getByRole('button', { name: 'Deny' })).toBeDisabled();
		expect(mocks.post).not.toHaveBeenCalled();
	});
});
