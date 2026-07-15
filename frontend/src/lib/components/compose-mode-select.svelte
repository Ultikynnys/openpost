<script lang="ts">
	import * as Select from '$lib/components/ui/select';
	import { cn } from '$lib/utils';
	import AlignLeftIcon from 'lucide-svelte/icons/align-left';
	import ImageIcon from 'lucide-svelte/icons/image';
	import ImagesIcon from 'lucide-svelte/icons/images';
	import LinkIcon from 'lucide-svelte/icons/link';
	import ListIcon from 'lucide-svelte/icons/list';
	import PlayIcon from 'lucide-svelte/icons/play';
	import SmartphoneIcon from 'lucide-svelte/icons/smartphone';
	import VideoIcon from 'lucide-svelte/icons/video';
	import {
		COMPOSER_MODE_GROUPS,
		COMPOSER_MODE_KEYS,
		composerMode,
		type ComposerModeKey
	} from './compose/modes';

	const modeIcons: Partial<Record<ComposerModeKey, typeof AlignLeftIcon>> = {
		short_text: AlignLeftIcon,
		thread: ListIcon,
		link_share: LinkIcon,
		image_post: ImageIcon,
		carousel: ImagesIcon,
		story: SmartphoneIcon,
		short_video: PlayIcon,
		long_video: VideoIcon
	};

	interface Props {
		selectedMode: ComposerModeKey;
		onModeChange: (mode: ComposerModeKey) => void;
		class?: string;
	}

	let { selectedMode, onModeChange, class: className }: Props = $props();
	const selectedModeMeta = $derived(composerMode(selectedMode));
	const SelectedIcon = $derived(modeIcons[selectedMode] ?? AlignLeftIcon);

	function handleValueChange(value: string) {
		if (!COMPOSER_MODE_KEYS.includes(value as ComposerModeKey)) return;
		onModeChange(value as ComposerModeKey);
	}
</script>

<Select.Root type="single" value={selectedMode} onValueChange={handleValueChange}>
	<Select.Trigger
		class={cn('h-8 w-32 max-w-32 min-w-0 text-xs', className)}
		aria-label="Post type"
		data-testid="composer-mode-select"
	>
		<span class="flex min-w-0 items-center gap-1.5">
			<SelectedIcon class="size-3.5 text-muted-foreground" />
			<span class="truncate">{selectedModeMeta.label}</span>
		</span>
	</Select.Trigger>
	<Select.Content
		class="max-h-[min(34rem,var(--bits-select-content-available-height))] w-80 max-w-[calc(100vw-1.5rem)]"
	>
		<Select.Label class="px-3 pt-2 pb-1 font-medium text-foreground">Post type</Select.Label>
		{#each COMPOSER_MODE_GROUPS as group, groupIndex (group.key)}
			{#if groupIndex > 0}
				<Select.Separator class="mx-2" />
			{/if}
			<Select.Group class="px-1.5 py-1">
				<Select.GroupHeading
					class="px-2 py-1 font-mono text-[0.625rem] font-medium tracking-[0.12em] uppercase"
				>
					{group.label}
				</Select.GroupHeading>
				{#each group.modes as mode (mode.key)}
					{@const Icon = modeIcons[mode.key] ?? AlignLeftIcon}
					<Select.Item
						value={mode.key}
						class="min-h-12 items-start gap-2.5 px-2 py-2 pr-8"
						data-testid={`composer-mode-option-${mode.key}`}
					>
						<Icon class="mt-0.5 size-4 text-muted-foreground" />
						<span class="min-w-0 flex-col items-start! gap-0!">
							<span class="text-xs/4 font-medium text-foreground">{mode.label}</span>
							<span class="line-clamp-2 text-[0.6875rem]/4 text-muted-foreground">
								{mode.description}
							</span>
						</span>
					</Select.Item>
				{/each}
			</Select.Group>
		{/each}
	</Select.Content>
</Select.Root>
