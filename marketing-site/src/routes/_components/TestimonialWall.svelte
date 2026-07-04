<script lang="ts">
	import { Button } from '$lib/components/ui/button';
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
	const visibleTestimonials = $derived(expanded ? testimonials : testimonials.slice(0, 4));
</script>

<section class="section-pad border-y bg-muted/25">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="max-w-3xl">
			<p class="eyebrow">Social proof shape</p>
			<h2 class="mt-3 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				{heading}
			</h2>
			<p class="mt-4 max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
				{description}
			</p>
		</div>

		<div class="mt-10 columns-1 gap-4 md:columns-2 xl:columns-3">
			{#each visibleTestimonials as testimonial (testimonial.id)}
				<article class="mb-4 break-inside-avoid rounded-lg border bg-card p-5 shadow-sm">
					<div class="flex items-center justify-between gap-4">
						<div class="flex items-center gap-3">
							<div
								class="flex size-10 items-center justify-center rounded-full border bg-muted font-mono text-xs font-semibold"
							>
								{testimonial.name
									.split(' ')
									.map((part) => part[0])
									.join('')
									.slice(0, 2)}
							</div>
							<div>
								<h3 class="text-sm font-semibold">{testimonial.name}</h3>
								<p class="text-xs text-muted-foreground">{testimonial.role}</p>
							</div>
						</div>
						<span class="rounded-full border px-2 py-1 text-xs text-muted-foreground">
							{testimonial.source}
						</span>
					</div>
					<p class="mt-4 text-sm leading-7 text-muted-foreground">{testimonial.content}</p>
				</article>
			{/each}
		</div>

		{#if testimonials.length > visibleTestimonials.length}
			<div class="mt-6 flex justify-center">
				<Button variant="secondary" onclick={() => (expanded = true)}>Show more notes</Button>
			</div>
		{/if}
	</div>
</section>
