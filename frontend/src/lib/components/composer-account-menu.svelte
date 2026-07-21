<script lang="ts">
	import type { SocialAccount } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import * as Popover from '$lib/components/ui/popover';
	import { cn, getPlatformKey, getPlatformName } from '$lib/utils';
	import { m } from '$lib/paraglide/messages';
	import CheckIcon from 'lucide-svelte/icons/check';
	import ChevronDownIcon from 'lucide-svelte/icons/chevron-down';
	import Link2Icon from 'lucide-svelte/icons/link-2';
	import PencilIcon from 'lucide-svelte/icons/pencil';
	import UnlinkIcon from 'lucide-svelte/icons/unlink';
	import PlatformIcon from './platform-icon.svelte';

	interface Props {
		accounts: SocialAccount[];
		selectedAccountIds: string[];
		compatibleAccountIds?: string[];
		customAccountIds?: string[];
		activeAccountId?: string | null;
		triggerLabel?: string;
		triggerClass?: string;
		triggerVariant?: 'ghost' | 'outline';
		description?: string;
		onToggle: (account: SocialAccount) => void;
		onSelectAll: () => void;
		onClearAll: () => void;
		onEditShared?: () => void;
		onCustomize?: (account: SocialAccount) => void;
		onReset?: (account: SocialAccount) => void;
	}

	let {
		accounts,
		selectedAccountIds,
		compatibleAccountIds,
		customAccountIds = [],
		activeAccountId = null,
		triggerLabel = m.compose_target_accounts(),
		triggerClass = '',
		triggerVariant = 'ghost',
		description = '',
		onToggle,
		onSelectAll,
		onClearAll,
		onEditShared,
		onCustomize,
		onReset
	}: Props = $props();

	let open = $state(false);
	const compatibleIds = $derived(
		new Set(compatibleAccountIds ?? accounts.map((account) => account.id))
	);
	const customIds = $derived(new Set(customAccountIds));
	const selectedAccounts = $derived(
		accounts.filter((account) => selectedAccountIds.includes(account.id))
	);
	const visibleAccounts = $derived(selectedAccounts.slice(0, 3));
	const hiddenAccountCount = $derived(
		Math.max(0, selectedAccounts.length - visibleAccounts.length)
	);
	const selectedSummary = $derived.by(() => {
		if (selectedAccounts.length === 0) return m.compose_no_accounts();
		if (selectedAccounts.length === compatibleIds.size) return m.compose_all_accounts();
		return m.compose_account_count({ count: selectedAccounts.length });
	});

	function accountLabel(account: SocialAccount): string {
		const username = account.account_username?.replace(/^@/, '');
		return `${getPlatformName(account.platform)}${username ? ` @${username}` : ''}`;
	}

	function editShared() {
		onEditShared?.();
		open = false;
	}

	function customize(account: SocialAccount) {
		onCustomize?.(account);
		open = false;
	}

	function reset(account: SocialAccount) {
		onReset?.(account);
		open = false;
	}
</script>

<Popover.Root bind:open>
	<Popover.Trigger>
		{#snippet child({ props })}
			<Button
				{...props}
				type="button"
				variant={triggerVariant}
				size="sm"
				class={cn('h-8 shrink-0 gap-1.5 px-2 text-xs text-muted-foreground', triggerClass)}
				aria-label={`${triggerLabel}: ${selectedSummary}`}
				data-testid="composer-account-control"
			>
				{#if selectedAccounts.length > 0}
					<span class="flex items-center -space-x-1.5" aria-hidden="true">
						{#each visibleAccounts as account (account.id)}
							<span
								class="flex size-6 items-center justify-center rounded-full border border-border bg-background ring-2 ring-background"
								data-testid="composer-account-icon"
							>
								<PlatformIcon platform={getPlatformKey(account.platform)} class="size-3.5" />
							</span>
						{/each}
						{#if hiddenAccountCount > 0}
							<span
								class="flex size-6 items-center justify-center rounded-full border border-border bg-muted text-[10px] font-medium text-muted-foreground ring-2 ring-background"
								aria-hidden="true"
							>
								+{hiddenAccountCount}
							</span>
						{/if}
					</span>
				{:else}
					<span>{m.common_none()}</span>
				{/if}
				<ChevronDownIcon class="size-3" aria-hidden="true" />
			</Button>
		{/snippet}
	</Popover.Trigger>

	<Popover.Content class="w-80 max-w-[calc(100vw-1rem)] p-2" align="start">
		<div class="flex items-start justify-between gap-3 px-2 py-1">
			<div class="min-w-0 py-1.5">
				<p class="text-sm font-medium">{m.compose_publish_to()}</p>
				{#if description}
					<p class="mt-0.5 text-xs leading-5 text-muted-foreground">{description}</p>
				{/if}
			</div>
			<div class="flex shrink-0 items-center gap-1">
				<Button type="button" variant="ghost" size="sm" class="min-h-11 px-2" onclick={onSelectAll}>
					{m.common_all()}
				</Button>
				<Button type="button" variant="ghost" size="sm" class="min-h-11 px-2" onclick={onClearAll}>
					{m.common_none()}
				</Button>
			</div>
		</div>

		{#if onEditShared}
			<div class="my-1 border-t pt-1">
				<button
					type="button"
					class={cn(
						'flex min-h-11 w-full items-center gap-2 rounded-md px-2 text-left text-sm transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
						activeAccountId === null && 'bg-accent/70'
					)}
					onclick={editShared}
					aria-pressed={activeAccountId === null}
					data-testid="composer-shared-content"
				>
					<Link2Icon class="size-4 text-muted-foreground" aria-hidden="true" />
					<span class="flex-1">{m.compose_shared_content()}</span>
					{#if activeAccountId === null}<CheckIcon class="size-4" aria-hidden="true" />{/if}
				</button>
			</div>
		{/if}

		<div class="space-y-1 border-t pt-1" role="group" aria-label={m.compose_publish_to()}>
			{#each accounts as account (account.id)}
				{@const compatible = compatibleIds.has(account.id)}
				{@const selected = selectedAccountIds.includes(account.id)}
				{@const custom = customIds.has(account.id)}
				<div
					class={cn(
						'flex min-h-11 items-stretch gap-1 rounded-md',
						activeAccountId === account.id && 'bg-accent/70'
					)}
					data-testid="composer-account-row"
				>
					<label
						class={cn(
							'flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md px-2 py-2 text-sm transition-colors focus-within:ring-2 focus-within:ring-ring hover:bg-accent',
							!compatible && 'cursor-not-allowed opacity-45'
						)}
					>
						<input
							type="checkbox"
							class="sr-only"
							checked={selected}
							disabled={!compatible}
							onchange={() => onToggle(account)}
						/>
						<span class="flex size-4 items-center justify-center" aria-hidden="true">
							<PlatformIcon platform={getPlatformKey(account.platform)} class="size-4" />
						</span>
						<span class="min-w-0 flex-1 truncate">{accountLabel(account)}</span>
						{#if custom}
							<span class="text-xs text-amber-600 dark:text-amber-400">{m.compose_custom()}</span>
						{/if}
						<span
							class={cn(
								'flex size-5 shrink-0 items-center justify-center rounded border',
								selected
									? 'border-primary bg-primary text-primary-foreground'
									: 'border-input bg-background'
							)}
							aria-hidden="true"
						>
							{#if selected}<CheckIcon class="size-3.5" />{/if}
						</span>
					</label>

					{#if compatible && selected && onCustomize}
						<Button
							type="button"
							variant="ghost"
							size="icon"
							class={cn('size-11 shrink-0', custom && 'text-amber-700 dark:text-amber-300')}
							title={custom
								? m.compose_edit_account_version({ account: accountLabel(account) })
								: m.compose_unsync()}
							aria-label={custom
								? m.compose_edit_account_version({ account: accountLabel(account) })
								: `${m.compose_unsync()}: ${accountLabel(account)}`}
							onclick={() => customize(account)}
							data-testid="composer-account-customize"
						>
							{#if custom}<PencilIcon class="size-4" />{:else}<UnlinkIcon class="size-4" />{/if}
						</Button>
					{/if}

					{#if compatible && selected && custom && onReset}
						<Button
							type="button"
							variant="ghost"
							size="icon"
							class="size-11 shrink-0"
							title={m.compose_sync_back()}
							aria-label={`${m.compose_sync_back()}: ${accountLabel(account)}`}
							onclick={() => reset(account)}
							data-testid="composer-account-reset"
						>
							<Link2Icon class="size-4" />
						</Button>
					{/if}
				</div>
			{/each}
		</div>
	</Popover.Content>
</Popover.Root>
