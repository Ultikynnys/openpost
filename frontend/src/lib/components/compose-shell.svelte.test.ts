import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ComposeShell from './compose-shell.svelte';

vi.mock('$lib/api/client', () => ({
	client: {
		GET: vi.fn(async () => ({ data: [], error: null })),
		POST: vi.fn(async () => ({ data: null, error: null })),
		PUT: vi.fn(async () => ({ data: null, error: null })),
		DELETE: vi.fn(async () => ({ data: null, error: null }))
	},
	getToken: () => null
}));

describe('ComposeShell', () => {
	it('defaults to the legacy composer shell instead of the unified rails', async () => {
		const screen = await render(ComposeShell);
		const picker = screen.getByTestId('composer-mode-picker');

		await expect.element(picker).toBeVisible();
		await expect.element(screen.getByTestId('legacy-composer-shell')).toBeVisible();
		await expect.element(picker.getByRole('button', { name: 'Post' })).toBeVisible();
		expect(screen.container.textContent).not.toContain('Format-first composer');
		expect(screen.container.textContent).not.toContain('Renditions');
		expect(screen.container.textContent).not.toContain('Source text');
	});
});
