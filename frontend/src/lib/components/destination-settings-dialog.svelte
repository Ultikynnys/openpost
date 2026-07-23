<script lang="ts">
	import type { SocialAccount } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { getPlatformName } from '$lib/utils';
	import { m } from '$lib/paraglide/messages';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import RotateCcwIcon from 'lucide-svelte/icons/rotate-ccw';
	import InlineNotice from './inline-notice.svelte';
	import PlatformIcon from './platform-icon.svelte';
	import TagInput from './tag-input.svelte';

	type SettingField = components['schemas']['SettingField'];
	type DestinationOption = components['schemas']['DestinationOption'];

	interface Props {
		open?: boolean;
		account: SocialAccount | null;
		settings: SettingField[];
		values: Record<string, unknown>;
		optionGroups?: Record<string, DestinationOption[]>;
		optionsLoading?: boolean;
		optionsError?: string;
		onChange: (key: string, value: unknown) => void;
		onRetry?: () => void;
	}

	let {
		open = $bindable(false),
		account,
		settings,
		values,
		optionGroups = {},
		optionsLoading = false,
		optionsError = '',
		onChange,
		onRetry
	}: Props = $props();

	function valueAsString(key: string): string {
		const value = values[key];
		return typeof value === 'string' || typeof value === 'number' ? String(value) : '';
	}

	function valueAsBoolean(key: string): boolean {
		return Boolean(values[key]);
	}

	function dynamicOptions(setting: SettingField): DestinationOption[] {
		if (!setting.options_source) return [];
		return optionGroups[setting.options_source] ?? [];
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-xl">
		<Dialog.Header>
			<Dialog.Title>
				<span class="flex items-center gap-2">
					{#if account}
						<PlatformIcon platform={account.platform} class="size-4" />
					{/if}
					{account
						? m.compose_account_settings({ platform: getPlatformName(account.platform) })
						: m.compose_platform_settings()}
				</span>
			</Dialog.Title>
			<Dialog.Description>
				{account?.account_username
					? m.compose_account_settings_body({ account: account.account_username })
					: m.compose_platform_settings_body()}
			</Dialog.Description>
		</Dialog.Header>

		{#if optionsError}
			<InlineNotice tone="error" message={optionsError}>
				{#snippet actions()}
					{#if onRetry}
						<Button type="button" variant="outline" size="sm" onclick={onRetry}>
							<RotateCcwIcon class="size-3.5" />
							{m.common_retry()}
						</Button>
					{/if}
				{/snippet}
			</InlineNotice>
		{/if}

		<div class="grid gap-4 py-1 sm:grid-cols-2">
			{#each settings as setting (setting.key)}
				{@const remoteOptions = dynamicOptions(setting)}
				<div class={setting.type === 'textarea' || setting.type === 'tags' ? 'sm:col-span-2' : ''}>
					{#if setting.type === 'boolean'}
						<label class="flex min-h-11 items-center gap-2 text-sm">
							<input
								type="checkbox"
								class="size-4 rounded border"
								checked={valueAsBoolean(setting.key)}
								onchange={(event) => onChange(setting.key, event.currentTarget.checked)}
							/>
							<span>{setting.label}</span>
						</label>
					{:else}
						<label class="text-sm font-medium" for="destination-setting-{setting.key}">
							{setting.label}
							{#if setting.required}
								<span class="text-destructive" aria-hidden="true">*</span>
							{/if}
						</label>
						{#if setting.type === 'select'}
							<select
								id="destination-setting-{setting.key}"
								class="mt-1 h-10 w-full rounded-md border bg-background px-2 text-sm"
								value={valueAsString(setting.key)}
								disabled={Boolean(setting.options_source) && optionsLoading}
								onchange={(event) => onChange(setting.key, event.currentTarget.value)}
							>
								{#if setting.options_source}
									<option value="" disabled={setting.required}>
										{setting.required
											? m.compose_choose_setting({ setting: setting.label })
											: m.common_none()}
									</option>
									{#each remoteOptions as option (option.value)}
										<option value={option.value}>{option.label}</option>
									{/each}
								{:else}
									{#each setting.options ?? [] as option (option)}
										<option value={option}>{option}</option>
									{/each}
								{/if}
							</select>
							{#if setting.options_source && optionsLoading}
								<p class="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
									<LoaderIcon class="size-3 animate-spin" />
									{m.compose_loading_provider_options()}
								</p>
							{:else if setting.options_source && !optionsError && remoteOptions.length === 0}
								<p class="mt-1 text-xs text-muted-foreground">
									{m.compose_no_provider_options({ setting: setting.label })}
								</p>
							{/if}
						{:else if setting.type === 'tags'}
							<TagInput
								id="destination-setting-{setting.key}"
								value={valueAsString(setting.key)}
								onChange={(value) => onChange(setting.key, value)}
							/>
						{:else if setting.type === 'textarea'}
							<Textarea
								id="destination-setting-{setting.key}"
								class="mt-1 min-h-24"
								value={valueAsString(setting.key)}
								oninput={(event) => onChange(setting.key, event.currentTarget.value)}
							/>
						{:else}
							<Input
								id="destination-setting-{setting.key}"
								class="mt-1"
								type={setting.type === 'number' ? 'number' : 'text'}
								value={valueAsString(setting.key)}
								oninput={(event) => onChange(setting.key, event.currentTarget.value)}
							/>
						{/if}
					{/if}
					{#if setting.help}
						<p class="mt-1 text-xs text-muted-foreground">{setting.help}</p>
					{/if}
				</div>
			{/each}
		</div>

		<Dialog.Footer>
			<Button type="button" onclick={() => (open = false)}>{m.common_done()}</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
