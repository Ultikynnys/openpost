import { describe, expect, it } from 'vitest';
import { videoFrameTimestamps, VIDEO_THUMBNAIL_FRAME_RATIOS } from './frames';

describe('videoFrameTimestamps', () => {
	it('spreads four frames evenly across a typical short video', () => {
		expect(videoFrameTimestamps(30_000)).toEqual([3000, 10500, 18000, 25500]);
		expect(VIDEO_THUMBNAIL_FRAME_RATIOS).toHaveLength(4);
	});

	it('clamps the last frame before the end of the video', () => {
		expect(videoFrameTimestamps(1_000)).toEqual([100, 350, 600, 750]);
	});

	it('dedupes colliding timestamps on very short videos', () => {
		const timestamps = videoFrameTimestamps(400);
		expect(new Set(timestamps).size).toBe(timestamps.length);
		expect(timestamps.length).toBeGreaterThan(0);
		expect(timestamps.length).toBeLessThanOrEqual(VIDEO_THUMBNAIL_FRAME_RATIOS.length);
	});

	it('returns no timestamps for invalid durations', () => {
		expect(videoFrameTimestamps(0)).toEqual([]);
		expect(videoFrameTimestamps(Number.NaN)).toEqual([]);
		expect(videoFrameTimestamps(Number.POSITIVE_INFINITY)).toEqual([]);
	});
});
