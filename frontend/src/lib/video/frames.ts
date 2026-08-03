/**
 * Client-side video frame extraction used by the composer's Thumbnail section
 * to offer thumbnails generated from the post's own video.
 *
 * Frames are captured from a paused <video> element into a 2D canvas and
 * encoded as JPEG blobs. The source is expected to be a same-origin
 * authenticated media URL so the canvas is never tainted.
 */

export const VIDEO_THUMBNAIL_FRAME_RATIOS = [0.1, 0.35, 0.6, 0.85] as const;

export interface ExtractedVideoFrame {
	/** JPEG blob ready to be uploaded as the thumbnail. */
	blob: Blob;
	/** Object URL for display. Callers must revoke it with URL.revokeObjectURL. */
	url: string;
	/** Frame position in the source video, in milliseconds. */
	timestampMs: number;
}

/**
 * Maps a video duration to evenly spaced frame timestamps, clamped so the last
 * frame is never the final frame of the video (often black) and deduplicated
 * for very short clips.
 */
export function videoFrameTimestamps(
	durationMs: number,
	ratios: readonly number[] = VIDEO_THUMBNAIL_FRAME_RATIOS
): number[] {
	if (!Number.isFinite(durationMs) || durationMs <= 0) return [];
	const lastSafeMs = Math.max(0, durationMs - 250);
	const timestamps = ratios.map((ratio) =>
		Math.min(lastSafeMs, Math.max(0, Math.round(durationMs * ratio)))
	);
	return [...new Set(timestamps)];
}

/**
 * Extracts one JPEG frame per ratio from the video at `source`, in order.
 * Throws if the video cannot be loaded, has no usable duration, or frame
 * encoding fails.
 */
export async function extractVideoFrames(
	source: string,
	ratios: readonly number[] = VIDEO_THUMBNAIL_FRAME_RATIOS
): Promise<ExtractedVideoFrame[]> {
	const video = document.createElement('video');
	video.preload = 'auto';
	video.muted = true;
	video.playsInline = true;
	video.src = source;
	const frames: ExtractedVideoFrame[] = [];
	try {
		await waitForVideoEvent(video, 'loadedmetadata');
		const timestamps = videoFrameTimestamps(video.duration * 1000, ratios);
		if (timestamps.length === 0) {
			throw new Error('Video has no usable duration');
		}
		const canvas = document.createElement('canvas');
		const context = canvas.getContext('2d', { alpha: false });
		if (!context) {
			throw new Error('Canvas 2D is unavailable');
		}
		canvas.width = video.videoWidth || 1280;
		canvas.height = video.videoHeight || 720;
		for (const timestampMs of timestamps) {
			video.currentTime = timestampMs / 1000;
			await waitForVideoEvent(video, 'seeked');
			// Let the browser composite the sought frame before drawing it.
			await nextAnimationFrame();
			await nextAnimationFrame();
			context.drawImage(video, 0, 0, canvas.width, canvas.height);
			const blob = await canvasToJpeg(canvas);
			frames.push({ blob, url: URL.createObjectURL(blob), timestampMs });
		}
		return frames;
	} finally {
		video.removeAttribute('src');
		video.load();
	}
}

function waitForVideoEvent(
	video: HTMLVideoElement,
	event: 'loadedmetadata' | 'seeked'
): Promise<void> {
	return new Promise((resolve, reject) => {
		const onReady = () => {
			cleanup();
			resolve();
		};
		const onError = () => {
			cleanup();
			reject(new Error(`Video failed while waiting for ${event}`));
		};
		const cleanup = () => {
			video.removeEventListener(event, onReady);
			video.removeEventListener('error', onError);
		};
		video.addEventListener(event, onReady, { once: true });
		video.addEventListener('error', onError, { once: true });
	});
}

function canvasToJpeg(canvas: HTMLCanvasElement): Promise<Blob> {
	return new Promise((resolve, reject) => {
		canvas.toBlob(
			(blob) => {
				if (blob) {
					resolve(blob);
				} else {
					reject(new Error('Could not encode video frame'));
				}
			},
			'image/jpeg',
			0.85
		);
	});
}

function nextAnimationFrame(): Promise<void> {
	return new Promise((resolve) => requestAnimationFrame(() => resolve()));
}
