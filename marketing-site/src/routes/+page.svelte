<script lang="ts">
	import { ArrowRight, ChevronRight, Command, Database, Terminal } from 'lucide-svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import FaqSection from './_components/FaqSection.svelte';
	import PricingComparison from './_components/PricingComparison.svelte';
	import VideoDemo from './_components/VideoDemo.svelte';
	import {
		appUrl,
		docsUrl,
		faqs,
		githubUrl,
		platforms,
		productFeatures,
		securityItems,
		selfHostingDocsUrl,
		siteUrl,
		workflowBlocks
	} from './_marketing';

	const heroFlow = [
		{ label: 'Draft', value: 'Base post' },
		{ label: 'Adapt', value: 'Per account' },
		{ label: 'Schedule', value: 'Next slot' },
		{ label: 'Publish', value: 'Visible status' }
	] as const;
</script>

<svelte:head>
	<title>OpenPost - Social publishing for creators and small teams</title>
	<meta
		name="description"
		content="OpenPost is a focused social publishing workspace for drafting, previewing, scheduling, and automating posts across modern platforms."
	/>
	<link rel="canonical" href={siteUrl} />
	<meta name="robots" content="index, follow" />
	<meta property="og:type" content="website" />
	<meta property="og:url" content={siteUrl} />
	<meta property="og:title" content="OpenPost - Social publishing workspace" />
	<meta
		property="og:description"
		content="Draft once, adapt per platform, schedule with confidence, and keep the queue visible."
	/>
	<meta property="og:image" content={`${siteUrl}/assets/brand/og-image.svg`} />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content="OpenPost - Social publishing workspace" />
	<meta
		name="twitter:description"
		content="Open-source social publishing for creators and small teams."
	/>
</svelte:head>

<section class="relative overflow-hidden border-b">
	<div class="marketing-grid pointer-events-none absolute inset-0"></div>
	<div class="relative mx-auto max-w-7xl px-4 py-16 sm:px-6 sm:py-20 lg:px-8 lg:py-24">
		<div class="grid items-center gap-12 lg:grid-cols-[0.88fr_1.12fr]">
			<div>
				<Badge class="border-primary/25 bg-primary/10 text-primary">
					Open-source publishing, ready to use
				</Badge>
				<h1
					class="mt-6 max-w-2xl text-4xl leading-[1.05] font-semibold tracking-tight text-balance sm:text-5xl"
				>
					A focused workspace for every post.
				</h1>
				<p class="mt-6 max-w-2xl text-lg leading-8 text-pretty text-muted-foreground sm:text-xl">
					Draft once, adapt for each account, preview the result, and keep every scheduled post
					visible from one place.
				</p>
				<div class="mt-8 flex flex-wrap gap-3">
					<Button href={appUrl} size="lg">
						Start publishing
						<ArrowRight data-icon="inline-end" />
					</Button>
					<Button href={selfHostingDocsUrl} variant="outline" size="lg">
						Self-host instead
						<ChevronRight data-icon="inline-end" />
					</Button>
				</div>
				<div class="mt-7 grid max-w-xl grid-cols-2 gap-3 text-sm sm:grid-cols-4">
					{#each heroFlow as step (step.label)}
						<div class="rounded-lg border bg-card/70 p-3">
							<p class="font-mono text-[0.68rem] tracking-[0.14em] text-muted-foreground uppercase">
								{step.label}
							</p>
							<p class="mt-1 font-medium">{step.value}</p>
						</div>
					{/each}
				</div>
			</div>

			<div class="relative">
				<div class="rounded-xl border bg-card p-2 shadow-[0_28px_90px_oklch(0.08_0.02_45/0.45)]">
					<div class="flex h-9 items-center gap-2 border-b px-3">
						<span class="size-2 rounded-full bg-[oklch(0.67_0.18_28)]"></span>
						<span class="size-2 rounded-full bg-[oklch(0.78_0.15_82)]"></span>
						<span class="size-2 rounded-full bg-[oklch(0.67_0.13_150)]"></span>
						<span class="ml-2 truncate font-mono text-xs text-muted-foreground">
							app.openpost.social
						</span>
					</div>
					<img
						src="/assets/screenshots/main-dark.png"
						alt="OpenPost composer and scheduling dashboard"
						class="mt-2 block w-full rounded-md border"
						width="1280"
						height="832"
						fetchpriority="high"
					/>
				</div>
			</div>
		</div>

		<div class="mt-12 border-y py-5">
			<div class="flex flex-col gap-5 lg:flex-row lg:items-center">
				<div class="flex flex-wrap gap-2">
					{#each platforms as platform (platform.slug)}
						<a
							href={`/platforms/${platform.slug}`}
							class="inline-flex items-center gap-2 rounded-full border bg-card px-3 py-2 text-sm font-medium transition hover:bg-muted"
						>
							<PlatformIcon platform={platform.short} class="size-4" />
							<span>{platform.name}</span>
						</a>
					{/each}
				</div>
			</div>
		</div>
	</div>
</section>
<VideoDemo />

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-10 lg:grid-cols-[0.78fr_1.22fr]">
			<div>
				<p class="eyebrow">Product workflow</p>
				<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
					Everything important happens before the post is sent.
				</h2>
				<p class="mt-5 max-w-xl text-lg leading-8 text-muted-foreground">
					OpenPost is not trying to become a giant social command center. It focuses on the work
					that makes publishing reliable: writing, adapting, previewing, scheduling, and seeing what
					happened.
				</p>
			</div>
			<div class="grid gap-4 sm:grid-cols-2">
				{#each workflowBlocks as block (block.title)}
					{@const Icon = block.icon}
					<article class="rounded-lg border bg-card p-5">
						<Icon class="size-5 text-primary" />
						<h3 class="mt-4 font-semibold">{block.title}</h3>
						<p class="mt-2 text-sm leading-6 text-muted-foreground">
							{block.description}
						</p>
					</article>
				{/each}
			</div>
		</div>

		<div class="mt-16 grid gap-6">
			{#each productFeatures as feature, index (feature.title)}
				{@const Icon = feature.icon}
				<article
					class="grid gap-6 rounded-xl border bg-card p-3 lg:grid-cols-[0.78fr_1.22fr] lg:items-center"
				>
					<div class="p-5 lg:p-8 {index % 2 === 1 ? 'lg:order-2' : ''}">
						<p class="eyebrow">{feature.eyebrow}</p>
						<h3 class="mt-4 text-2xl leading-tight font-semibold text-balance sm:text-4xl">
							{feature.title}
						</h3>
						<p class="mt-4 text-base leading-7 text-muted-foreground">
							{feature.description}
						</p>
						<div
							class="mt-6 inline-flex items-center gap-2 rounded-full border px-3 py-2 text-sm text-muted-foreground"
						>
							<Icon class="size-4 text-primary" />
							<span>{feature.eyebrow} workflow</span>
						</div>
					</div>
					<figure class="overflow-hidden rounded-lg border bg-background">
						<img
							src={feature.image}
							alt={feature.alt}
							class="block aspect-[16/10] w-full object-cover object-top"
							width="1280"
							height="832"
							loading="lazy"
						/>
					</figure>
				</article>
			{/each}
		</div>
	</div>
</section>

<section class="section-pad border-y bg-muted/25">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-10 lg:grid-cols-[1fr_1fr] lg:items-center">
			<div>
				<p class="eyebrow">Automation</p>
				<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
					Web app for editorial work. CLI and MCP for everything around it.
				</h2>
				<p class="mt-5 text-lg leading-8 text-muted-foreground">
					Use the app to review content, then let scripts, CI, cron, or assistants create drafts,
					upload media, select accounts, and schedule approved posts.
				</p>
				<div class="mt-8 flex flex-wrap gap-3">
					<Button href="/tools" variant="secondary">
						Explore free tools
						<ArrowRight data-icon="inline-end" />
					</Button>
					<Button href={docsUrl} variant="outline">Read automation docs</Button>
				</div>
			</div>
			<div class="grid gap-4">
				<div class="rounded-lg border bg-card p-5 font-mono text-sm leading-7">
					<p class="text-muted-foreground">$ openpost auth login https://app.openpost.social</p>
					<p class="text-muted-foreground">
						$ openpost post create --accounts x,linkedin --content 'Launch day.' --schedule
						next-slot
					</p>
					<p class="mt-3 text-primary">scheduled for the next workspace slot</p>
				</div>
				<div class="grid gap-4 sm:grid-cols-3">
					{#each [Terminal, Command, Database] as Icon, i (`automation-${i}`)}
						<div class="rounded-lg border bg-card p-4">
							<Icon class="size-5 text-primary" />
							<p class="mt-3 text-sm font-medium">
								{['CLI', 'MCP', 'API'][i]}
							</p>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
</section>

<PricingComparison />

<section class="section-pad">
	<div class="mx-auto grid max-w-7xl gap-10 px-4 sm:px-6 lg:grid-cols-[0.85fr_1.15fr] lg:px-8">
		<div>
			<p class="eyebrow">Trust</p>
			<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				Security and openness are part of the product, not a footnote.
			</h2>
			<p class="mt-5 text-lg leading-8 text-muted-foreground">
				OpenPost handles provider tokens and automation credentials with the same practical
				discipline it applies to scheduling.
			</p>
			<Button href="/security" class="mt-8" variant="secondary">Security details</Button>
		</div>
		<div class="grid gap-4 sm:grid-cols-2">
			{#each securityItems as item (item.title)}
				{@const Icon = item.icon}
				<article class="rounded-lg border bg-card p-5">
					<Icon class="size-5 text-primary" />
					<h3 class="mt-4 font-semibold">{item.title}</h3>
					<p class="mt-2 text-sm leading-6 text-muted-foreground">
						{item.description}
					</p>
				</article>
			{/each}
		</div>
	</div>
</section>

<section class="border-y bg-primary text-primary-foreground">
	<div
		class="mx-auto grid max-w-7xl gap-8 px-4 py-12 sm:px-6 lg:grid-cols-[1fr_auto] lg:items-center lg:px-8"
	>
		<div>
			<h2 class="text-3xl font-semibold text-balance">Prefer to run it yourself?</h2>
			<p class="mt-3 max-w-2xl text-base leading-7 text-primary-foreground/75">
				OpenPost is open source, self-hostable, and intentionally lightweight: one binary or
				container, SQLite by default, local media by default, and no Redis queue requirement.
			</p>
		</div>
		<div class="flex flex-wrap gap-3">
			<Button href={selfHostingDocsUrl} variant="secondary">Self-hosting docs</Button>
			<Button
				href={githubUrl}
				variant="outline"
				class="border-primary-foreground/25 text-primary-foreground hover:bg-primary-foreground/10"
			>
				View GitHub
			</Button>
		</div>
	</div>
</section>

<FaqSection items={faqs} />

<section class="section-pad border-t">
	<div class="mx-auto max-w-5xl px-4 text-center sm:px-6 lg:px-8">
		<p class="eyebrow">Start focused</p>
		<h2 class="mt-4 text-4xl leading-tight font-semibold text-balance sm:text-6xl">
			Queue the next post without adopting a social media suite.
		</h2>
		<p class="mx-auto mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">
			Open the app, create the first workspace, connect the platforms that matter, and keep
			publishing visible.
		</p>
		<div class="mt-8 flex justify-center gap-3">
			<Button href={appUrl} size="lg">
				Open app
				<ArrowRight data-icon="inline-end" />
			</Button>
			<Button href="/#pricing" variant="outline" size="lg">See plans</Button>
		</div>
	</div>
</section>
