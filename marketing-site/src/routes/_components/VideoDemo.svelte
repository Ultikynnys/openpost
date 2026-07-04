<script lang="ts">
	import { Play, X } from 'lucide-svelte';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		videoSrc?: string | null;
		thumbnailSrc?: string;
		thumbnailAlt?: string;
	}

	let {
		videoSrc = null,
		thumbnailSrc = '/assets/screenshots/main-dark.png',
		thumbnailAlt = 'OpenPost composer video thumbnail'
	}: Props = $props();

	let open = $state(false);

	function close() {
		open = false;
	}

	function handleModalClick(event: MouseEvent) {
		if (event.target === event.currentTarget) close();
	}

	function handleWindowKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && open) close();
	}
</script>

<svelte:window onkeydown={handleWindowKeydown} />

<section class="section-pad border-y bg-muted/20">
	<div class="mx-auto grid max-w-7xl gap-10 px-4 sm:px-6 lg:grid-cols-[0.72fr_1.28fr] lg:items-center lg:px-8">
		<div>
			<Badge class="border-primary/25 bg-background text-muted-foreground">Product demo</Badge>
			<h2 class="mt-5 max-w-2xl text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				See the composer, variants, previews, and queue in one pass.
			</h2>
			<p class="mt-5 max-w-xl text-lg leading-8 text-muted-foreground">
				This space is ready for the recorded tour: draft once, adapt per destination, check limits,
				attach media, and schedule from the same workflow.
			</p>
		</div>

		<button
			type="button"
			class="video-trigger group"
			aria-label="Play OpenPost product video"
			onclick={() => (open = true)}
		>
			<img src={thumbnailSrc} alt={thumbnailAlt} loading="lazy" decoding="async" />
			<span class="video-surface" aria-hidden="true">
				<span class="video-play-ring">
					<span class="video-play">
						<Play class="size-8 fill-current" />
					</span>
				</span>
			</span>
		</button>
	</div>
</section>

{#if open}
	<div class="video-modal" role="presentation" onclick={handleModalClick}>
		<div class="video-dialog" role="dialog" aria-modal="true" aria-label="OpenPost product video">
			<button type="button" class="video-close" aria-label="Close video" onclick={close}>
				<X class="size-5" />
			</button>
			{#if videoSrc}
				<iframe
					src={videoSrc}
					title="OpenPost product video"
					allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
					allowfullscreen
				></iframe>
			{:else}
				<div class="video-placeholder">
					<img src={thumbnailSrc} alt="" />
					<div>
						<p>OpenPost product demo</p>
						<span>Recording slot ready</span>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	.video-trigger {
		position: relative;
		display: block;
		overflow: hidden;
		width: 100%;
		border: 1px solid var(--border);
		border-radius: 0.9rem;
		background: var(--muted);
		padding: 0;
		box-shadow: 0 24px 70px color-mix(in oklch, black 24%, transparent);
		cursor: pointer;
	}

	.video-trigger img {
		display: block;
		width: 100%;
		aspect-ratio: 16 / 9;
		object-fit: cover;
		object-position: top;
		transition:
			filter 180ms ease,
			transform 240ms ease;
	}

	.video-trigger:hover img {
		filter: brightness(0.72);
		transform: scale(1.015);
	}

	.video-surface {
		position: absolute;
		inset: 0;
		display: grid;
		place-items: center;
		background: color-mix(in oklch, black 20%, transparent);
	}

	.video-play-ring {
		display: grid;
		width: clamp(5rem, 11vw, 7rem);
		height: clamp(5rem, 11vw, 7rem);
		place-items: center;
		border-radius: 999px;
		background: color-mix(in oklch, var(--background) 22%, transparent);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		transition: transform 180ms ease;
	}

	.video-trigger:hover .video-play-ring {
		transform: scale(1.04);
	}

	.video-play {
		display: grid;
		width: 4.2rem;
		height: 4.2rem;
		place-items: center;
		border-radius: 999px;
		background: var(--primary);
		color: var(--primary-foreground);
		box-shadow: 0 14px 28px color-mix(in oklch, black 28%, transparent);
	}

	.video-modal {
		position: fixed;
		z-index: 220;
		inset: 0;
		display: grid;
		place-items: center;
		background: color-mix(in oklch, black 64%, transparent);
		padding: 1rem;
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		animation: video-modal-in 140ms ease both;
	}

	.video-dialog {
		position: relative;
		width: min(66rem, 100%);
		aspect-ratio: 16 / 9;
		border: 1px solid color-mix(in oklch, white 30%, transparent);
		border-radius: 0.9rem;
		background: black;
		box-shadow: 0 28px 90px color-mix(in oklch, black 34%, transparent);
		animation: video-dialog-in 180ms ease both;
	}

	.video-dialog iframe {
		width: 100%;
		height: 100%;
		border: 0;
		border-radius: inherit;
	}

	.video-placeholder {
		position: relative;
		width: 100%;
		height: 100%;
		overflow: hidden;
		border-radius: inherit;
	}

	.video-placeholder img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: top;
		filter: brightness(0.48);
	}

	.video-placeholder div {
		position: absolute;
		inset: 0;
		display: grid;
		place-content: center;
		text-align: center;
		color: white;
	}

	.video-placeholder p {
		font-size: clamp(1.6rem, 4vw, 3rem);
		font-weight: 650;
	}

	.video-placeholder span {
		margin-top: 0.75rem;
		font-size: 0.9rem;
		color: color-mix(in oklch, white 72%, transparent);
	}

	.video-close {
		position: absolute;
		top: -3.1rem;
		right: 0;
		display: grid;
		width: 2.4rem;
		height: 2.4rem;
		place-items: center;
		border: 1px solid color-mix(in oklch, white 20%, transparent);
		border-radius: 999px;
		background: color-mix(in oklch, black 42%, transparent);
		color: white;
	}

	@keyframes video-modal-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@keyframes video-dialog-in {
		from {
			opacity: 0;
			transform: translateY(1rem) scale(0.98);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.video-trigger img,
		.video-play-ring,
		.video-modal,
		.video-dialog {
			animation: none;
			transition: none;
		}
	}
</style>
