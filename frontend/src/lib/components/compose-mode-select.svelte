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
		COMPOSER_MODE_KEYS,
		SELECTABLE_COMPOSER_MODES,
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
		class={cn('h-8 max-w-44 min-w-34 justify-start text-xs sm:min-w-40', className)}
		aria-label="Post type"
		data-testid="composer-mode-select"
	>
		<SelectedIcon class="size-3.5 text-muted-foreground" />
		<span class="truncate">{selectedModeMeta.label}</span>
	</Select.Trigger>
	<Select.Content class="w-56">
		<Select.Label>Post type</Select.Label>
		{#each SELECTABLE_COMPOSER_MODES as mode (mode.key)}
			{@const Icon = modeIcons[mode.key] ?? AlignLeftIcon}
			<Select.Item value={mode.key}>
				<Icon class="size-3.5 text-muted-foreground" />
				<span>{mode.label}</span>
			</Select.Item>
		{/each}
	</Select.Content>
</Select.Root>
