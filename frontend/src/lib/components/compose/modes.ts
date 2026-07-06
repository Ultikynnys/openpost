import { getPlatformKey, getPlatformName } from '$lib/utils';

export const COMPOSER_MODE_KEYS = [
	'short_text',
	'thread',
	'link_share',
	'image_post',
	'carousel',
	'story',
	'short_video',
	'long_video'
] as const;

export type ComposerModeKey = (typeof COMPOSER_MODE_KEYS)[number];

export type FocusedFieldKey =
	'postText' | 'caption' | 'linkUrl' | 'videoTitle' | 'videoDescription';

export type FocusedFieldType = 'text' | 'textarea' | 'url';

export interface ComposerMode {
	key: ComposerModeKey;
	label: string;
	description: string;
	flow: 'legacy' | 'publication';
	mediaFirst: boolean;
}

export interface ComposerAccountTarget {
	id: string;
	platform: string;
	account_username?: string;
}

export interface FocusedRoleField {
	key: FocusedFieldKey;
	label: string;
	hint: string;
	type: FocusedFieldType;
	required?: boolean;
	rows?: number;
}

export interface FocusedComposerFields {
	postText?: string;
	caption?: string;
	linkUrl?: string;
	videoTitle?: string;
	videoDescription?: string;
}

export interface FocusedPublicationInput {
	mode: ComposerModeKey;
	workspaceId: string;
	accounts: ComposerAccountTarget[];
	fields: FocusedComposerFields;
	mediaIds: string[];
	scheduledAt?: string;
	thumbnailMediaId?: string;
	settingsByAccount?: Record<string, Record<string, unknown>>;
}

export interface FocusedPublicationPayload {
	workspace_id: string;
	title: string;
	content_profile: ComposerModeKey;
	source_text: string;
	source_url?: string;
	scheduled_at?: string;
	metadata: Record<string, unknown>;
	media: Array<{ media_id: string; role: string }>;
	renditions: Array<{
		social_account_id: string;
		profile: ComposerModeKey;
		body: string;
		title: string;
		description: string;
		settings: Record<string, unknown>;
		media: Array<{ media_id: string; role: string }>;
	}>;
}

export const COMPOSER_MODES: ComposerMode[] = [
	{
		key: 'short_text',
		label: 'Short text',
		description: 'Fast text-first posts for timelines and feeds.',
		flow: 'legacy',
		mediaFirst: false
	},
	{
		key: 'thread',
		label: 'Thread',
		description: 'Use the existing add-post flow for ordered replies.',
		flow: 'legacy',
		mediaFirst: false
	},
	{
		key: 'link_share',
		label: 'Link',
		description: 'A URL-led post with clear post text.',
		flow: 'publication',
		mediaFirst: false
	},
	{
		key: 'image_post',
		label: 'Image',
		description: 'Single image or simple media feed post.',
		flow: 'publication',
		mediaFirst: true
	},
	{
		key: 'carousel',
		label: 'Carousel',
		description: 'Multiple images or document-style swipes.',
		flow: 'publication',
		mediaFirst: true
	},
	{
		key: 'story',
		label: 'Story',
		description: 'Ephemeral vertical story media.',
		flow: 'publication',
		mediaFirst: true
	},
	{
		key: 'short_video',
		label: 'Short video',
		description: 'Reels, Shorts, TikTok, and short-form video.',
		flow: 'publication',
		mediaFirst: true
	},
	{
		key: 'long_video',
		label: 'Long video',
		description: 'YouTube and feed video uploads with metadata.',
		flow: 'publication',
		mediaFirst: false
	}
];

export function composerMode(key: ComposerModeKey): ComposerMode {
	return COMPOSER_MODES.find((mode) => mode.key === key) ?? COMPOSER_MODES[0];
}

export function isLegacyComposerMode(key: ComposerModeKey): boolean {
	return composerMode(key).flow === 'legacy';
}

export function roleFieldsForMode(
	mode: ComposerModeKey,
	accounts: ComposerAccountTarget[]
): FocusedRoleField[] {
	const platforms = unique(accounts.map((account) => getPlatformKey(account.platform)));
	const hasYouTube = platforms.includes('youtube');
	const nonYouTubePlatforms = platforms.filter((platform) => platform !== 'youtube');
	const nonYouTubeHint = platformHint(nonYouTubePlatforms);

	switch (mode) {
		case 'short_text':
		case 'thread':
			return [
				{
					key: 'postText',
					label: 'Post text',
					hint: 'Post text',
					type: 'textarea',
					required: true,
					rows: 8
				}
			];
		case 'link_share':
			return [
				{
					key: 'linkUrl',
					label: 'Link URL',
					hint: 'Shared link',
					type: 'url',
					required: true
				},
				{
					key: 'postText',
					label: 'Post text',
					hint: 'Post text',
					type: 'textarea',
					rows: 7
				}
			];
		case 'image_post':
		case 'carousel':
			return [captionField(platformHint(platforms), true)];
		case 'story':
			return [];
		case 'short_video': {
			const fields: FocusedRoleField[] = [];
			if (hasYouTube) fields.push(videoTitleField(), videoDescriptionField());
			if (!hasYouTube || nonYouTubePlatforms.length > 0) {
				fields.push(captionField(nonYouTubeHint || 'TikTok, Instagram, Facebook, Threads'));
			}
			return fields;
		}
		case 'long_video': {
			const fields = [videoTitleField(), videoDescriptionField()];
			if (nonYouTubePlatforms.length > 0) fields.push(captionField(nonYouTubeHint));
			return fields;
		}
	}
}

export function buildFocusedPublicationPayload(
	input: FocusedPublicationInput
): FocusedPublicationPayload {
	const media = input.mediaIds.map((mediaId, index) => ({
		media_id: mediaId,
		role: 'attachment'
	}));
	const sourceText = firstNonEmpty(
		input.fields.videoDescription,
		input.fields.caption,
		input.fields.postText,
		input.fields.videoTitle,
		input.fields.linkUrl
	);
	const title = firstNonEmpty(
		input.fields.videoTitle,
		firstLine(input.fields.caption),
		firstLine(input.fields.postText),
		firstLine(input.fields.videoDescription),
		composerMode(input.mode).label
	);

	return {
		workspace_id: input.workspaceId,
		title,
		content_profile: input.mode,
		source_text: sourceText,
		...(input.fields.linkUrl?.trim() ? { source_url: input.fields.linkUrl.trim() } : {}),
		...(input.scheduledAt ? { scheduled_at: input.scheduledAt } : {}),
		metadata: {
			composer: 'focused',
			mode: input.mode
		},
		media,
		renditions: input.accounts.map((account) =>
			renditionPayload({
				account,
				input,
				defaultTitle: title,
				defaultSourceText: sourceText,
				media
			})
		)
	};
}

function renditionPayload({
	account,
	input,
	defaultTitle,
	defaultSourceText,
	media
}: {
	account: ComposerAccountTarget;
	input: FocusedPublicationInput;
	defaultTitle: string;
	defaultSourceText: string;
	media: FocusedPublicationPayload['media'];
}): FocusedPublicationPayload['renditions'][number] {
	const platform = getPlatformKey(account.platform);
	const settings = {
		...(input.settingsByAccount?.[account.id] ?? {})
	};

	if (input.fields.linkUrl?.trim()) {
		settings.link_url ??= input.fields.linkUrl.trim();
		settings.url ??= input.fields.linkUrl.trim();
	}

	if (platform === 'youtube') {
		settings.privacy ??= 'private';
		if (input.fields.videoTitle?.trim()) settings.title ??= input.fields.videoTitle.trim();
		if (input.fields.videoDescription?.trim()) {
			settings.description ??= input.fields.videoDescription.trim();
		}
		if (input.thumbnailMediaId) settings.thumbnail_media_id ??= input.thumbnailMediaId;
	}

	const youtubeBody = firstNonEmpty(
		input.fields.videoDescription,
		input.fields.caption,
		input.fields.postText,
		input.fields.videoTitle
	);
	const socialBody = firstNonEmpty(
		input.fields.caption,
		input.fields.postText,
		input.fields.videoDescription,
		input.fields.videoTitle,
		defaultSourceText
	);
	const title =
		platform === 'youtube' ? firstNonEmpty(input.fields.videoTitle, defaultTitle) : defaultTitle;
	const description =
		platform === 'youtube'
			? firstNonEmpty(input.fields.videoDescription)
			: firstNonEmpty(input.fields.videoDescription, input.fields.caption);

	return {
		social_account_id: account.id,
		profile: input.mode,
		body: platform === 'youtube' ? youtubeBody : socialBody,
		title,
		description,
		settings,
		media
	};
}

function captionField(hint: string, required = false): FocusedRoleField {
	return {
		key: 'caption',
		label: 'Caption',
		hint: hint ? `Caption · ${hint}` : 'Caption',
		type: 'textarea',
		required,
		rows: 7
	};
}

function videoTitleField(): FocusedRoleField {
	return {
		key: 'videoTitle',
		label: 'Video title',
		hint: 'Video title · YouTube title',
		type: 'text',
		required: true
	};
}

function videoDescriptionField(): FocusedRoleField {
	return {
		key: 'videoDescription',
		label: 'Description',
		hint: 'Description · YouTube description',
		type: 'textarea',
		rows: 8
	};
}

function platformHint(platforms: string[]): string {
	return unique(platforms).map(getPlatformName).join(', ');
}

function unique(values: string[]): string[] {
	return Array.from(new Set(values.filter(Boolean)));
}

function firstLine(value: string | undefined): string {
	return value?.trim().split('\n').find(Boolean)?.trim() ?? '';
}

function firstNonEmpty(...values: Array<string | undefined>): string {
	return values.map((value) => value?.trim() ?? '').find(Boolean) ?? '';
}
