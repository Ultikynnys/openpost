import type { SocialAccount } from '$lib/api/client';
import { getPlatformKey, getPlatformName } from '$lib/utils';

export const DEFAULT_PLATFORM_CHAR_LIMIT = 280;

export const PLATFORM_CHAR_LIMITS: Record<string, number> = {
	x: 280,
	mastodon: 500,
	bluesky: 300,
	linkedin: 3000,
	threads: 500
};

export interface PlatformLimit {
	platform: string;
	key: string;
	limit: number;
}

export function accountCharacterLimit(account: Pick<SocialAccount, 'platform'>): number {
	return PLATFORM_CHAR_LIMITS[getPlatformKey(account.platform)] ?? DEFAULT_PLATFORM_CHAR_LIMIT;
}

export function minimumAccountCharacterLimit(
	accounts: Array<Pick<SocialAccount, 'platform'>>
): number {
	if (accounts.length === 0) return DEFAULT_PLATFORM_CHAR_LIMIT;
	return Math.min(...accounts.map(accountCharacterLimit));
}

export function uniquePlatformLimits(
	accounts: Array<Pick<SocialAccount, 'platform'>>
): PlatformLimit[] {
	const seen = new Set<string>();
	return accounts
		.map((account) => {
			const key = getPlatformKey(account.platform);
			return {
				platform: getPlatformName(account.platform),
				key,
				limit: PLATFORM_CHAR_LIMITS[key] ?? DEFAULT_PLATFORM_CHAR_LIMIT
			};
		})
		.filter((item) => {
			if (seen.has(item.key)) return false;
			seen.add(item.key);
			return true;
		});
}
