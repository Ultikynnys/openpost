<script lang="ts">
	import { Github } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import type { testimonials as defaultTestimonials } from '../_marketing';

	type Testimonial = (typeof defaultTestimonials)[number];

	interface Props {
		testimonials: readonly Testimonial[];
		heading?: string;
		description?: string;
	}

	let {
		testimonials,
		heading = 'Built for the people who actually operate the queue',
		description = 'Representative notes from the creators, small teams, and technical operators OpenPost is designed around.'
	}: Props = $props();

	let expanded = $state(false);
	const visibleTestimonials = $derived(expanded ? testimonials : testimonials.slice(0, 5));

	function sourcePlatform(source: string) {
		const normalized = source.toLowerCase();
		if (normalized === 'x') return 'x';
		if (normalized === 'linkedin') return 'linkedin';
		if (normalized === 'mastodon') return 'mastodon';
		if (normalized === 'bluesky') return 'bluesky';
		return null;
	}
</script>

<section class="section-pad border-y bg-muted/25">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="max-w-3xl">
			<p class="eyebrow">Testimonials</p>
			<h2 class="mt-3 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				{heading}
			</h2>
			<p class="mt-4 max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
				{description}
			</p>
		</div>

		<div class="relative mt-10">
			<div class="columns-1 gap-4 md:columns-2 xl:columns-3">
				{#each visibleTestimonials as testimonial (testimonial.id)}
					{@const platform = sourcePlatform(testimonial.source)}
					<article
						class="mb-4 break-inside-avoid rounded-lg border bg-card p-5 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
					>
						<div class="flex items-center justify-between gap-4">
							<div class="flex min-w-0 items-center gap-3">
								<img
									src={testimonial.avatar}
									alt=""
									class="size-11 shrink-0 rounded-full border object-cover"
									loading="lazy"
									decoding="async"
								/>
								<div class="min-w-0">
									<h3 class="truncate text-sm font-semibold">{testimonial.name}</h3>
									<p class="truncate text-xs text-muted-foreground">{testimonial.role}</p>
								</div>
							</div>
							<span
								class="inline-flex shrink-0 items-center gap-1.5 rounded-full border px-2 py-1 text-xs text-muted-foreground"
							>
								{#if platform}
									<PlatformIcon platform={platform} class="size-3.5" />
								{:else}
									<Github class="size-3.5" />
								{/if}
								{testimonial.source}
							</span>
						</div>
						<p class="mt-4 text-sm leading-7 text-muted-foreground">{testimonial.content}</p>
					</article>
				{/each}
			</div>

			{#if !expanded && testimonials.length > visibleTestimonials.length}
				<div
					class="pointer-events-none absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-muted/25 to-transparent"
				></div>
			{/if}
		</div>

		{#if testimonials.length > visibleTestimonials.length}
			<div class="mt-6 flex justify-center">
				<Button variant="secondary" onclick={() => (expanded = !expanded)}>
					{expanded ? 'Show fewer notes' : 'Show more notes'}
				</Button>
			</div>
		{/if}
	</div>
</section>
