<script lang="ts">
	import {
		ArrowRight,
		CalendarClock,
		Command,
		Library,
		MessageSquareText,
		ShieldCheck,
		Workflow
	} from 'lucide-svelte';
	import PageHero from '../_components/PageHero.svelte';
	import { comparisons, siteUrl } from '../_marketing';

	const featureHighlights = [
		{
			title: 'Multi-platform composer',
			description:
				'Draft the base post once, then split copy and media per destination when limits or tone differ.',
			icon: MessageSquareText
		},
		{
			title: 'Scheduling and queue visibility',
			description:
				'Use workspace posting slots, next-slot scheduling, job states, failures, and activity history.',
			icon: CalendarClock
		},
		{
			title: 'Workspaces and account destinations',
			description:
				'Separate brands, clients, products, schedules, prompts, media, and connected accounts.',
			icon: Workflow
		},
		{
			title: 'Reusable media library',
			description:
				'Keep assets, alt text, usage, favorites, and deletion safety checks attached to publishing work.',
			icon: Library
		},
		{
			title: 'CLI, MCP, and API automation',
			description:
				'Let scripts, CI, cron, and assistants create drafts, upload media, and schedule approved posts.',
			icon: Command
		},
		{
			title: 'Open-source trust',
			description:
				'Inspect the implementation, use the managed app, or self-host the lightweight single-service stack.',
			icon: ShieldCheck
		}
	] as const;
</script>

<svelte:head>
	<title>Compare OpenPost - OpenPost</title>
	<meta
		name="description"
		content="Compare OpenPost with Buffer, Hootsuite, Typefully, Postiz, Post Bridge, and Mixpost."
	/>
	<link rel="canonical" href={`${siteUrl}/compare`} />
</svelte:head>

<PageHero
	eyebrow="Compare"
	title="A comparison hub that does not hide the tradeoffs."
	description="OpenPost is not the biggest social suite. It is the focused, open-source publishing workspace for people who care about queue visibility, automation, and lightweight operation."
	secondaryHref="/#pricing"
	secondaryLabel="See pricing"
/>

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="max-w-3xl">
			<p class="eyebrow">OpenPost strengths</p>
			<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				What you get before comparing vendors.
			</h2>
			<p class="mt-5 text-lg leading-8 text-muted-foreground">
				The comparison pages start from the real product surface: composer, variants, queue,
				media, automation, and open-source operation.
			</p>
		</div>

		<div class="mt-10 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each featureHighlights as feature (feature.title)}
				{@const Icon = feature.icon}
				<article class="rounded-xl border bg-card p-5">
					<Icon class="size-5 text-primary" />
					<h3 class="mt-4 font-semibold">{feature.title}</h3>
					<p class="mt-2 text-sm leading-6 text-muted-foreground">{feature.description}</p>
				</article>
			{/each}
		</div>

		<div class="mt-12 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each comparisons as comparison (comparison.slug)}
				<a href={`/compare/${comparison.slug}`} class="group rounded-xl border bg-card p-5 transition hover:bg-muted/30">
					<div class="flex items-center justify-between gap-4">
						<h2 class="font-semibold">OpenPost vs {comparison.name}</h2>
						<ArrowRight class="size-4 text-muted-foreground transition group-hover:translate-x-0.5 group-hover:text-foreground" />
					</div>
					<p class="mt-4 text-sm leading-6 text-muted-foreground">{comparison.openPostAngle}</p>
				</a>
			{/each}
		</div>
	</div>
</section>
