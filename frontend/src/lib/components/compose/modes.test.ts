import { describe, expect, it } from 'vitest';
import {
	buildFocusedPublicationPayload,
	COMPOSER_MODE_GROUPS,
	roleFieldsForMode,
	SELECTABLE_COMPOSER_MODES
} from './modes';

const youtube = { id: 'yt-1', platform: 'youtube', account_username: 'OpenPost' };
const tiktok = { id: 'tt-1', platform: 'tiktok', account_username: 'openpost' };
const instagram = { id: 'ig-1', platform: 'instagram', account_username: 'openpost' };

describe('composer mode role mapping', () => {
	it('keeps the format picker aligned with user intent instead of media API shapes', () => {
		expect(SELECTABLE_COMPOSER_MODES.map((mode) => mode.label)).toEqual([
			'Post',
			'Thread',
			'Link',
			'Image',
			'Carousel',
			'Story',
			'Short video',
			'Video'
		]);
	});

	it('groups every selectable format once without hiding secondary types', () => {
		expect(
			COMPOSER_MODE_GROUPS.map((group) => ({
				label: group.label,
				modes: group.modes.map((mode) => mode.key)
			}))
		).toEqual([
			{ label: 'Write', modes: ['short_text', 'thread', 'link_share'] },
			{
				label: 'Media',
				modes: ['image_post', 'carousel', 'story', 'short_video', 'long_video']
			}
		]);
		expect(COMPOSER_MODE_GROUPS.flatMap((group) => group.modes)).toEqual(SELECTABLE_COMPOSER_MODES);
	});

	it('uses post text for short text', () => {
		expect(roleFieldsForMode('short_text', [])).toEqual([
			expect.objectContaining({ key: 'postText', label: 'Post text' })
		]);
	});

	it('uses link URL plus post text for link shares', () => {
		expect(roleFieldsForMode('link_share', [])).toEqual([
			expect.objectContaining({ key: 'linkUrl', label: 'Link URL' }),
			expect.objectContaining({ key: 'postText', label: 'Post text' })
		]);
	});

	it('uses caption for image posts', () => {
		expect(roleFieldsForMode('image_post', [instagram])).toEqual([
			expect.objectContaining({ key: 'caption', label: 'Caption', hint: 'Caption · Instagram' })
		]);
	});

	it('separates YouTube video metadata from social captions', () => {
		expect(roleFieldsForMode('short_video', [youtube, tiktok])).toEqual([
			expect.objectContaining({ key: 'videoTitle', label: 'Video title' }),
			expect.objectContaining({ key: 'videoDescription', label: 'Video description' }),
			expect.objectContaining({ key: 'caption', label: 'Caption', hint: 'Caption · TikTok' })
		]);
	});

	it('keeps long video YouTube-style first and adds caption for feed video targets', () => {
		expect(roleFieldsForMode('long_video', [youtube, instagram])).toEqual([
			expect.objectContaining({ key: 'videoTitle', label: 'Video title' }),
			expect.objectContaining({ key: 'videoDescription', label: 'Video description' }),
			expect.objectContaining({ key: 'caption', label: 'Caption', hint: 'Caption · Instagram' })
		]);
	});
});

describe('focused publication payloads', () => {
	it('maps YouTube title and description into rendition title and description', () => {
		const payload = buildFocusedPublicationPayload({
			mode: 'long_video',
			workspaceId: 'ws-1',
			accounts: [youtube],
			fields: {
				videoTitle: 'Launch walkthrough',
				videoDescription: 'A complete tour of the release.',
				caption: 'Watch the new release.'
			},
			mediaIds: ['video-1'],
			thumbnailMediaId: 'thumb-1'
		});

		expect(payload.source_text).toBe('A complete tour of the release.');
		expect(payload.title).toBe('Launch walkthrough');
		expect(payload.renditions).toHaveLength(1);
		expect(payload.renditions[0]).toMatchObject({
			social_account_id: 'yt-1',
			profile: 'long_video',
			body: 'A complete tour of the release.',
			title: 'Launch walkthrough',
			description: 'A complete tour of the release.',
			settings: {
				privacy: 'private',
				title: 'Launch walkthrough',
				description: 'A complete tour of the release.',
				thumbnail_media_id: 'thumb-1'
			}
		});
	});

	it('maps mixed short-video targets without overloading one textarea', () => {
		const payload = buildFocusedPublicationPayload({
			mode: 'short_video',
			workspaceId: 'ws-1',
			accounts: [youtube, tiktok],
			fields: {
				videoTitle: 'Launch in 60 seconds',
				videoDescription: 'The YouTube description.',
				caption: 'The TikTok caption.'
			},
			mediaIds: ['video-1']
		});

		expect(payload.renditions).toEqual([
			expect.objectContaining({
				social_account_id: 'yt-1',
				body: 'The YouTube description.',
				title: 'Launch in 60 seconds',
				description: 'The YouTube description.'
			}),
			expect.objectContaining({
				social_account_id: 'tt-1',
				body: 'The TikTok caption.',
				title: 'Launch in 60 seconds',
				description: 'The YouTube description.'
			})
		]);
	});
});
