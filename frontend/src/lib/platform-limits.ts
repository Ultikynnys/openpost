import { getPlatformKey, getPlatformName } from '$lib/utils';

export const DEFAULT_PLATFORM_CHAR_LIMIT = 280;
export const X_STANDARD_CHAR_LIMIT = 280;
export const X_PREMIUM_CHAR_LIMIT = 25_000;

export type AccountLimitProfile = 'standard' | 'x-premium';

export interface PlatformLimitDefinition {
	key: string;
	name: string;
	charLimit: number;
	media: string;
	note: string;
}

export interface PlatformLimit {
	platform: string;
	key: string;
	limit: number;
	profile?: AccountLimitProfile;
	note?: string;
}

export const PLATFORM_LIMITS: Record<string, PlatformLimitDefinition> = {
	x: {
		key: 'x',
		name: 'X',
		charLimit: X_STANDARD_CHAR_LIMIT,
		media: 'Up to 4 images or 1 MP4 video',
		note: 'Standard X posts use the 280-character API limit. X Premium longer posts can be up to 25,000 characters, but scheduled/API behavior should be verified per account.'
	},
	mastodon: {
		key: 'mastodon',
		name: 'Mastodon',
		charLimit: 500,
		media: 'Up to 4 attachments',
		note: 'Instance rules can vary.'
	},
	bluesky: {
		key: 'bluesky',
		name: 'Bluesky',
		charLimit: 300,
		media: 'Up to 4 images or 1 MP4 video',
		note: 'Video is MP4-only and cannot be mixed with images.'
	},
	linkedin: {
		key: 'linkedin',
		name: 'LinkedIn',
		charLimit: 3000,
		media: 'OpenPost currently publishes the first attachment',
		note: 'Thread children publish as comments when configured.'
	},
	threads: {
		key: 'threads',
		name: 'Threads',
		charLimit: 500,
		media: 'Single-media composer path',
		note: 'Media must be served from a public URL.'
	},
	facebook: {
		key: 'facebook',
		name: 'Facebook Pages',
		charLimit: 63206,
		media: 'One image or video URL',
		note: 'Pages publishing depends on Meta permissions and app review.'
	},
	instagram: {
		key: 'instagram',
		name: 'Instagram Business',
		charLimit: 2200,
		media: 'Exactly one image or video',
		note: 'Business accounts behind Facebook Pages only.'
	},
	tiktok: {
		key: 'tiktok',
		name: 'TikTok',
		charLimit: 2200,
		media: 'Exactly one video',
		note: 'Provider review may apply.'
	},
	youtube: {
		key: 'youtube',
		name: 'YouTube',
		charLimit: 5000,
		media: 'Exactly one video',
		note: 'Current adapter path uploads scheduled videos as private.'
	}
};

export function accountLimitProfile(account: {
	platform: string;
	limit_profile?: string | null;
}): AccountLimitProfile {
	if (getPlatformKey(account.platform) === 'x' && account.limit_profile === 'x-premium') {
		return 'x-premium';
	}
	return 'standard';
}

export function platformCharacterLimit(
	platform: string,
	profile: AccountLimitProfile = 'standard'
): number {
	if (getPlatformKey(platform) === 'x' && profile === 'x-premium') return X_PREMIUM_CHAR_LIMIT;
	return PLATFORM_LIMITS[getPlatformKey(platform)]?.charLimit ?? DEFAULT_PLATFORM_CHAR_LIMIT;
}

export function accountCharacterLimit(account: {
	platform: string;
	limit_profile?: string | null;
}) {
	return platformCharacterLimit(account.platform, accountLimitProfile(account));
}

export function minimumAccountCharacterLimit(
	accounts: Array<{ platform: string; limit_profile?: string | null }>
): number {
	if (accounts.length === 0) return DEFAULT_PLATFORM_CHAR_LIMIT;
	return Math.min(...accounts.map(accountCharacterLimit));
}

export function uniquePlatformLimits(
	accounts: Array<{ platform: string; limit_profile?: string | null }>
): PlatformLimit[] {
	const seen = new Set<string>();
	return accounts
		.map((account) => {
			const key = getPlatformKey(account.platform);
			const profile = accountLimitProfile(account);
			return {
				platform: profile === 'x-premium' ? 'X Premium' : getPlatformName(account.platform),
				key,
				profile,
				limit: platformCharacterLimit(account.platform, profile),
				note: PLATFORM_LIMITS[key]?.note
			};
		})
		.filter((item) => {
			const dedupeKey = `${item.key}:${item.profile}`;
			if (seen.has(dedupeKey)) return false;
			seen.add(dedupeKey);
			return true;
		});
}

export function publicPlatformLimits(): PlatformLimitDefinition[] {
	return Object.values(PLATFORM_LIMITS);
}
