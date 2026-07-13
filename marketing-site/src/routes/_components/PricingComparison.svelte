<script lang="ts">
	import { Check, X } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { appUrl, plans } from '../_marketing';

	const featureGroups = [
		{
			title: 'Plan limits',
			features: [
				{ title: 'Workspaces', values: ['1', '3', '10', '10', '50'] },
				{ title: 'Social accounts', values: ['3', '6', '15', '25', '150'] },
				{ title: 'Scheduled posts/month', values: ['100', '500', '2,500', '5,000', '25,000'] },
				{ title: 'Media storage', values: ['1 GB', '5 GB', '25 GB', 'Team pool', 'Agency pool'] },
				{ title: 'Included seats', values: ['1', '1', '1', '3', '5'] }
			]
		},
		{
			title: 'Publishing workflow',
			features: [
				{ title: 'Composer and drafts', values: ['check', 'check', 'check', 'check', 'check'] },
				{ title: 'Platform variants', values: ['check', 'check', 'check', 'check', 'check'] },
				{ title: 'Media library', values: ['check', 'check', 'check', 'check', 'check'] },
				{ title: 'Account destinations', values: ['check', 'check', 'check', 'check', 'check'] },
				{ title: 'Queue and activity states', values: ['check', 'check', 'check', 'check', 'check'] }
			]
		},
		{
			title: 'Automation and trust',
			features: [
				{ title: 'CLI', values: ['check', 'check', 'check', 'check', 'check'] },
				{ title: 'MCP tools', values: ['check', 'check', 'check', 'check', 'check'] },
				{ title: 'API tokens', values: ['check', 'check', 'check', 'check', 'check'] },
				{ title: 'Encrypted provider tokens', values: ['check', 'check', 'check', 'check', 'check'] },
				{ title: 'Team roles', values: ['cross', 'cross', 'cross', 'check', 'check'] }
			]
		}
	] as const;
</script>

<section id="pricing" class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="max-w-3xl">
			<p class="eyebrow">Pricing</p>
			<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				Plans that scale by publishing volume, not by confusion.
			</h2>
			<p class="mt-5 text-lg leading-8 text-muted-foreground">
				Start small, then add workspaces, accounts, seats, and monthly scheduled-post capacity as
				the publishing operation grows.
			</p>
		</div>

		<div class="mt-10 overflow-x-auto rounded-xl border bg-card">
			<div class="min-w-[920px]">
				<div class="grid grid-cols-[1.3fr_repeat(5,1fr)] gap-4 border-b p-4">
					<div class="flex items-end">
						<span class="text-sm font-medium text-muted-foreground">Monthly</span>
					</div>
					{#each plans as plan (plan.id)}
						<div class="rounded-lg border bg-background/50 p-3 {plan.featured ? 'border-primary/60' : ''}">
							<div class="flex items-center justify-between gap-2">
								<h3 class="font-semibold">{plan.name}</h3>
								{#if plan.featured}
									<span class="rounded-full bg-primary px-2 py-1 text-[0.65rem] text-primary-foreground">
										Popular
									</span>
								{/if}
							</div>
							<p class="mt-2 text-3xl font-semibold">
								{plan.price}<span class="text-sm text-muted-foreground">/mo</span>
							</p>
							<Button
								href={`${appUrl}/register?plan=${plan.id}`}
								class="mt-4 w-full"
								variant={plan.featured ? 'default' : 'outline'}
							>
								Choose
							</Button>
						</div>
					{/each}
				</div>

				{#each featureGroups as group (group.title)}
					<div class="border-b last:border-b-0">
						<h3 class="bg-muted/30 px-4 py-3 text-sm font-semibold">{group.title}</h3>
						{#each group.features as feature (feature.title)}
							<div class="grid grid-cols-[1.3fr_repeat(5,1fr)] gap-4 border-t px-4 py-3 text-sm">
								<div class="font-medium">{feature.title}</div>
								{#each feature.values as value, index (`${feature.title}-${index}`)}
									<div class="text-muted-foreground">
										{#if value === 'check'}
											<Check class="size-4 text-primary" />
										{:else if value === 'cross'}
											<X class="size-4 text-muted-foreground/60" />
										{:else}
											{value}
										{/if}
									</div>
								{/each}
							</div>
						{/each}
					</div>
				{/each}
			</div>
		</div>
	</div>
</section>
