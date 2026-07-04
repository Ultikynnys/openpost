<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { appUrl } from '../_marketing';

	interface Props {
		eyebrow: string;
		title: string;
		description: string;
		secondaryHref?: string;
		secondaryLabel?: string;
	}

	let {
		eyebrow,
		title,
		description,
		secondaryHref = '/pricing',
		secondaryLabel = 'See pricing'
	}: Props = $props();
</script>

<section class="border-b py-14 sm:py-16 lg:py-20">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="max-w-3xl">
			<p class="eyebrow">{eyebrow}</p>
			<h1 class="mt-4 text-4xl leading-[1.03] font-semibold text-balance sm:text-6xl">
				{title}
			</h1>
			<p class="mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">{description}</p>
			<div class="mt-8 flex flex-wrap gap-3">
				<Button href={appUrl} size="lg">Open app</Button>
				<Button href={secondaryHref} variant="outline" size="lg">{secondaryLabel}</Button>
			</div>
		</div>
	</div>
</section>
