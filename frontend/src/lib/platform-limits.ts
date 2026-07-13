import { getPlatformKey, getPlatformName } from '$lib/utils';

export const DEFAULT_PLATFORM_CHAR_LIMIT = 280;
export const X_STANDARD_CHAR_LIMIT = 280;
export const X_PREMIUM_CHAR_LIMIT = 25_000;

export type AccountLimitProfile = 'standard' | 'x-premium';

type AccountLimitTarget = {
	platform: string;
	limit_profile?: string | null;
	capabilities?: unknown;
	metadata?: unknown;
	[key: string]: unknown;
};

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
		media: 'One image, video, or document per rendition',
		note: 'Thread children publish as comments; use format-first publications for video and documents.'
	},
	threads: {
		key: 'threads',
		name: 'Threads',
		charLimit: 500,
		media: 'One media item or a 2-10 item carousel',
		note: 'Media must be served from public HTTPS URLs.'
	},
	facebook: {
		key: 'facebook',
		name: 'Facebook Pages',
		charLimit: 63206,
		media: 'One image/video, a 2-10 photo post, or one Story item',
		note: 'Pages publishing depends on Meta permissions and app review.'
	},
	instagram: {
		key: 'instagram',
		name: 'Instagram Business',
		charLimit: 2200,
		media: 'One image/video or a 2-10 item carousel',
		note: 'Business accounts behind Facebook Pages only.'
	},
	tiktok: {
		key: 'tiktok',
		name: 'TikTok',
		charLimit: 2200,
		media: 'One video or 1-35 JPEG/WebP photos',
		note: 'Public URL ownership verification and provider review may apply.'
	},
	youtube: {
		key: 'youtube',
		name: 'YouTube',
		charLimit: 5000,
		media: 'Exactly one video',
		note: 'Current adapter path uploads scheduled videos as private.'
	}
};

function hasPremiumToken(value: unknown): boolean {
	if (typeof value !== 'string') return false;
	const normalized = value.toLowerCase().replace(/[\s-]+/g, '_');
	return ['x_premium', 'premium', 'premium_plus', 'long_posts', 'longform'].includes(normalized);
}

function hasPremiumFlag(value: unknown): boolean {
	if (!value || typeof value !== 'object') return false;
	return Object.entries(value as Record<string, unknown>).some(([key, raw]) => {
		const normalizedKey = key.toLowerCase().replace(/[\s-]+/g, '_');
		if (
			['x_premium', 'has_x_premium', 'premium', 'premium_plus', 'long_posts', 'longform'].includes(
				normalizedKey
			)
		) {
			return raw === true || hasPremiumToken(raw);
		}
		if (
			['tier', 'profile', 'plan', 'subscription_tier', 'account_type', 'entitlement'].includes(
				normalizedKey
			)
		) {
			return hasPremiumToken(raw);
		}
		if (
			Array.isArray(raw) &&
			['capabilities', 'features', 'entitlements', 'permissions'].includes(normalizedKey)
		) {
			return raw.some(hasPremiumToken);
		}
		if (normalizedKey === 'metadata' || normalizedKey === 'settings') {
			return hasPremiumFlag(raw);
		}
		return false;
	});
}

export function accountHasXPremiumLongPosts(account: AccountLimitTarget): boolean {
	if (getPlatformKey(account.platform) !== 'x') return false;
	if (account.limit_profile === 'x-premium') return true;

	const capabilityValues = Array.isArray(account.capabilities) ? account.capabilities : [];
	if (capabilityValues.some(hasPremiumToken)) return true;

	if (hasPremiumFlag(account.metadata)) return true;
	if (hasPremiumFlag(account)) return true;

	return false;
}

export function accountLimitProfile(account: AccountLimitTarget): AccountLimitProfile {
	if (accountHasXPremiumLongPosts(account)) return 'x-premium';
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
	capabilities?: unknown;
	metadata?: unknown;
	[key: string]: unknown;
}) {
	return platformCharacterLimit(account.platform, accountLimitProfile(account));
}

export function minimumAccountCharacterLimit(accounts: Array<AccountLimitTarget>): number {
	if (accounts.length === 0) return DEFAULT_PLATFORM_CHAR_LIMIT;
	return Math.min(...accounts.map(accountCharacterLimit));
}

export function uniquePlatformLimits(accounts: Array<AccountLimitTarget>): PlatformLimit[] {
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
