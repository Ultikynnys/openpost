<script lang="ts">
	import { page } from '$app/state';
	import { error } from '@sveltejs/kit';
	import {
		ArrowRight,
		CalendarClock,
		CheckCircle2,
		Command,
		Library,
		MessageSquareText,
		ShieldCheck,
		Workflow
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import { appUrl, getComparison, platforms, siteUrl } from '../../_marketing';

	const slug = $derived(page.params.slug ?? '');
	const comparison = $derived.by(() => {
		const found = getComparison(slug);
		if (!found) error(404, 'Comparison not found');
		return found;
	});

	const comparisonRows = [
		{
			feature: 'Composer and platform variants',
			openpost: 'Base posts, per-account variants, previews, and provider warnings.',
			other: 'Varies by product; often tied to a broader suite workflow.',
			icon: MessageSquareText
		},
		{
			feature: 'Scheduling workflow',
			openpost: 'Workspace slots, next-slot scheduling, drafts, jobs, failures, and activity state.',
			other: 'Scheduling is usually present, but queue visibility and failure states vary.',
			icon: CalendarClock
		},
		{
			feature: 'Workspace operations',
			openpost: 'Workspaces, account destinations, media reuse, prompts, roles on team plans, and usage limits.',
			other: 'May be stronger for enterprise governance, approvals, analytics, or engagement.',
			icon: Workflow
		},
		{
			feature: 'Media library',
			openpost: 'Reusable assets, alt text, usage tracking, favorites, and delete safety checks.',
			other: 'Some tools optimize for quick upload instead of a reusable publishing library.',
			icon: Library
		},
		{
			feature: 'Automation',
			openpost: 'Web app plus CLI, MCP tools, API tokens, and assistant-friendly workflows.',
			other: 'Many tools expose APIs, but fewer center operator automation as a first-class path.',
			icon: Command
		},
		{
			feature: 'Trust model',
			openpost: 'Source-open implementation, encrypted provider tokens, and a lightweight self-host path.',
			other: 'Hosted-only suites can still be excellent, but the implementation is usually closed.',
			icon: ShieldCheck
		}
	] as const;
</script>

<svelte:head>
	<title>OpenPost vs {comparison.name} - OpenPost</title>
	<meta
		name="description"
		content={`Compare OpenPost with ${comparison.name} for social scheduling and publishing.`}
	/>
	<link rel="canonical" href={`${siteUrl}/compare/${comparison.slug}`} />
</svelte:head>

<section class="section-pad border-b">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-10 lg:grid-cols-[0.82fr_1.18fr] lg:items-center">
			<div>
			<p class="eyebrow">Comparison</p>
			<h1 class="mt-4 text-4xl leading-tight font-semibold text-balance sm:text-6xl">
				OpenPost vs {comparison.name}
			</h1>
			<p class="mt-5 text-lg leading-8 text-muted-foreground">{comparison.openPostAngle}</p>
			<p class="mt-4 text-base leading-7 text-muted-foreground">
				{comparison.name} may be better for: {comparison.bestFor}
			</p>
			<div class="mt-8 flex flex-wrap gap-3">
				<Button href={appUrl} size="lg">Open app</Button>
				<Button href="/compare" variant="outline" size="lg">All comparisons</Button>
			</div>
			</div>
			<div class="grid gap-4">
				<article class="rounded-xl border bg-card p-5">
					<h2 class="font-semibold">Choose OpenPost when</h2>
					<ul class="mt-4 grid gap-3 text-sm leading-6 text-muted-foreground">
						{#each comparison.chooseOpenPost as item (item)}
							<li class="flex gap-2">
								<CheckCircle2 class="mt-0.5 size-4 shrink-0 text-primary" />
								<span>{item}</span>
							</li>
						{/each}
					</ul>
				</article>
				<article class="rounded-xl border bg-muted/25 p-5">
					<h2 class="font-semibold">Choose {comparison.name} when</h2>
					<ul class="mt-4 grid gap-3 text-sm leading-6 text-muted-foreground">
						{#each comparison.chooseThem as item (item)}
							<li class="flex gap-2">
								<ArrowRight class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
								<span>{item}</span>
							</li>
						{/each}
					</ul>
				</article>
			</div>
		</div>
	</div>
</section>

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="max-w-3xl">
			<p class="eyebrow">Feature comparison</p>
			<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				OpenPost is strongest when publishing operations matter.
			</h2>
			<p class="mt-5 text-lg leading-8 text-muted-foreground">
				The table below focuses on features OpenPost actually builds around, while keeping the
				competitor column as fit guidance rather than a checklist fight.
			</p>
		</div>

		<div class="mt-10 overflow-hidden rounded-xl border bg-card">
			<div class="hidden grid-cols-[0.9fr_1.1fr_1.1fr] gap-4 border-b bg-muted/25 p-4 text-sm font-semibold md:grid">
				<span>Area</span>
				<span>OpenPost</span>
				<span>{comparison.name}</span>
			</div>
			{#each comparisonRows as row (row.feature)}
				{@const Icon = row.icon}
				<div class="grid gap-4 border-b p-5 last:border-b-0 md:grid-cols-[0.9fr_1.1fr_1.1fr]">
					<div class="flex items-center gap-3 font-semibold">
						<Icon class="size-5 text-primary" />
						{row.feature}
					</div>
					<p class="text-sm leading-6 text-muted-foreground">{row.openpost}</p>
					<p class="text-sm leading-6 text-muted-foreground">{row.other}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<section class="section-pad border-y bg-muted/25">
	<div class="mx-auto grid max-w-7xl gap-10 px-4 sm:px-6 lg:grid-cols-[0.78fr_1.22fr] lg:px-8">
		<div>
			<p class="eyebrow">Channels</p>
			<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				The platform map is broad, but still described carefully.
			</h2>
			<p class="mt-5 text-lg leading-8 text-muted-foreground">
				OpenPost exposes provider-specific limits and caveats in the publishing workflow instead
				of pretending every network behaves the same.
			</p>
		</div>
		<div class="grid gap-3 sm:grid-cols-2">
			{#each platforms.slice(0, 9) as platform (platform.slug)}
				<a
					href={`/platforms/${platform.slug}`}
					class="flex items-start gap-3 rounded-lg border bg-card p-4 transition hover:bg-background"
				>
					<PlatformIcon platform={platform.short} class="mt-0.5 size-5" />
					<span>
						<span class="block font-medium">{platform.name}</span>
						<span class="mt-1 block text-sm leading-6 text-muted-foreground">{platform.tag}</span>
					</span>
				</a>
			{/each}
		</div>
	</div>
</section>

<section class="section-pad">
	<div class="mx-auto grid max-w-7xl gap-8 px-4 sm:px-6 lg:grid-cols-[1fr_1fr] lg:px-8">
		<div class="rounded-xl border bg-card p-6">
			<p class="eyebrow">Pricing</p>
			<h2 class="mt-4 text-2xl leading-tight font-semibold text-balance">
				Plan limits that are easy to read.
			</h2>
			<p class="mt-4 text-sm leading-7 text-muted-foreground">{comparison.pricing}</p>
			<Button href="/#pricing" class="mt-6" variant="secondary">Compare OpenPost plans</Button>
		</div>
		<div class="rounded-xl border bg-card p-6">
			<p class="eyebrow">Positioning</p>
			<h2 class="mt-4 text-2xl leading-tight font-semibold text-balance">
				Focused scheduler, not a social suite replacement.
			</h2>
			<p class="mt-4 text-sm leading-7 text-muted-foreground">
				OpenPost centers the publishing loop: write, adapt, preview, schedule, monitor, and
				automate. If your buying requirement is advanced analytics, social listening, or
				enterprise approval workflows, another suite may be the better fit.
			</p>
		</div>
	</div>
</section>

<section class="section-pad border-t">
	<div class="mx-auto max-w-5xl px-4 text-center sm:px-6 lg:px-8">
		<p class="eyebrow">Try the focused path</p>
		<h2 class="mt-4 text-4xl leading-tight font-semibold text-balance sm:text-6xl">
			Start with the composer, queue, media, and automation.
		</h2>
		<p class="mx-auto mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">
			OpenPost is a better fit when the team wants a clean publishing workspace instead of a
			large social media suite.
		</p>
		<div class="mt-8 flex flex-wrap justify-center gap-3">
			<Button href={appUrl} size="lg">
				Open app
				<ArrowRight data-icon="inline-end" />
			</Button>
			<Button href="/tools" variant="outline" size="lg">Try free tools</Button>
		</div>
	</div>
</section>
