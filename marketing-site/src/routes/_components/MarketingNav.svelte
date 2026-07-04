<script lang="ts">
	import { ArrowRight, Github, Menu } from 'lucide-svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { Button } from '$lib/components/ui/button';
	import { appUrl, githubUrl, navItems } from '../_marketing';

	let mobileOpen = $state(false);
</script>

<header class="sticky top-0 z-40 border-b border-border/80 bg-background/92 backdrop-blur-xl">
	<div class="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-3 sm:px-6 lg:px-8">
		<a class="inline-flex items-center gap-2" href="/" aria-label="OpenPost home">
			<Logo width={34} height={26} />
			<span class="text-sm font-semibold tracking-tight">OpenPost</span>
		</a>

		<nav class="hidden items-center gap-1 md:flex" aria-label="Primary navigation">
			{#each navItems as item (item.href)}
				<a
					href={item.href}
					class="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition hover:bg-muted/60 hover:text-foreground"
				>
					{item.label}
				</a>
			{/each}
		</nav>

		<div class="hidden items-center gap-2 md:flex">
			<Button href={githubUrl} variant="ghost" size="sm" target="_blank" rel="noreferrer">
				<Github data-icon="inline-start" />
				Source
			</Button>
			<Button href={appUrl} size="sm">
				Open app
				<ArrowRight data-icon="inline-end" />
			</Button>
		</div>

		<Button
			variant="ghost"
			size="icon-sm"
			class="md:hidden"
			aria-label="Toggle navigation"
			aria-expanded={mobileOpen}
			onclick={() => (mobileOpen = !mobileOpen)}
		>
			<Menu />
		</Button>
	</div>

	{#if mobileOpen}
		<nav class="border-t bg-background px-4 py-4 md:hidden" aria-label="Mobile navigation">
			<div class="grid gap-1">
				{#each navItems as item (item.href)}
					<a
						href={item.href}
						class="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground"
						onclick={() => (mobileOpen = false)}
					>
						{item.label}
					</a>
				{/each}
			</div>
			<div class="mt-4 grid grid-cols-2 gap-2">
				<Button href={githubUrl} variant="outline" size="sm" target="_blank" rel="noreferrer">
					<Github data-icon="inline-start" />
					Source
				</Button>
				<Button href={appUrl} size="sm">
					Open app
					<ArrowRight data-icon="inline-end" />
				</Button>
			</div>
		</nav>
	{/if}
</header>
