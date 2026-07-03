import { describe, expect, it } from 'vitest';
import {
	DEFAULT_PLATFORM_CHAR_LIMIT,
	accountCharacterLimit,
	minimumAccountCharacterLimit,
	uniquePlatformLimits
} from './platform-limits';

describe('platform-limits', () => {
	it('uses canonical limits for known platforms', () => {
		expect(accountCharacterLimit({ platform: 'x' })).toBe(280);
		expect(accountCharacterLimit({ platform: 'mastodon:https://masto.pt' })).toBe(500);
		expect(accountCharacterLimit({ platform: 'linkedin' })).toBe(3000);
	});

	it('falls back to the default limit for unknown providers', () => {
		expect(accountCharacterLimit({ platform: 'unknown' })).toBe(DEFAULT_PLATFORM_CHAR_LIMIT);
	});

	it('returns the tightest selected account limit', () => {
		expect(
			minimumAccountCharacterLimit([
				{ platform: 'linkedin' },
				{ platform: 'bluesky' },
				{ platform: 'threads' }
			])
		).toBe(300);
	});

	it('deduplicates displayed platform limits by canonical platform', () => {
		expect(
			uniquePlatformLimits([
				{ platform: 'x' },
				{ platform: 'twitter' },
				{ platform: 'mastodon:https://masto.pt' }
			])
		).toEqual([
			{ platform: 'X', key: 'x', limit: 280 },
			{ platform: 'Mastodon', key: 'mastodon', limit: 500 }
		]);
	});
});
