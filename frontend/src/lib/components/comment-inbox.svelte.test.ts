import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import CommentInbox from './comment-inbox.svelte';
import type { components } from '$lib/api/types';

type CommentResponse = components['schemas']['CommentResponse'];

const baseComment: CommentResponse = {
	id: 'comment-1',
	provider_comment_id: 'provider-comment-1',
	rendition_id: 'rendition-1',
	text: 'This launch looks useful.',
	author_name: 'Rita',
	can_reply: true,
	can_hide: false,
	can_delete: false,
	hidden: false
};

describe('CommentInbox', () => {
	it('loads comments for a rendition and hides unsupported actions', async () => {
		const fakeClient = {
			listComments: vi.fn(async () => ({
				comments: [baseComment]
			})),
			replyToComment: vi.fn(),
			hideComment: vi.fn(),
			deleteComment: vi.fn()
		};

		const screen = await render(CommentInbox, {
			renditionId: 'rendition-1',
			platform: 'linkedin',
			client: fakeClient
		});

		await expect.element(screen.getByText('This launch looks useful.')).toBeVisible();
		await expect.element(screen.getByRole('button', { name: 'Reply' })).toBeVisible();
		await expect.element(screen.getByRole('button', { name: 'Hide' })).not.toBeInTheDocument();
		await expect.element(screen.getByRole('button', { name: 'Delete' })).not.toBeInTheDocument();
		expect(fakeClient.listComments).toHaveBeenCalledWith('rendition-1');
	});
});
