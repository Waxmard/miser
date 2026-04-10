<script lang="ts">
	import * as si from 'simple-icons';
	import type { SimpleIcon } from 'simple-icons';
	import { createEventDispatcher, onMount } from 'svelte';

	export let current: string | null = null; // current slug
	export let open = false;

	const dispatch = createEventDispatcher<{ select: string | null; close: void }>();

	const allIcons: SimpleIcon[] = Object.values(si).sort((a, b) =>
		a.title.localeCompare(b.title)
	);

	let query = '';
	let inputEl: HTMLInputElement;

	$: filtered =
		query.trim() === ''
			? allIcons
			: allIcons.filter(
					(icon) =>
						icon.title.toLowerCase().includes(query.toLowerCase()) ||
						icon.slug.includes(query.toLowerCase())
				);

	// Show first 120 results to avoid rendering all 3500+ at once
	$: visible = filtered.slice(0, 120);

	function select(slug: string) {
		dispatch('select', slug);
		open = false;
		query = '';
	}

	function clear() {
		dispatch('select', null);
		open = false;
		query = '';
	}

	function close() {
		dispatch('close');
		open = false;
		query = '';
	}

	function isUsable(hex: string): boolean {
		const r = parseInt(hex.slice(0, 2), 16) / 255;
		const g = parseInt(hex.slice(2, 4), 16) / 255;
		const b = parseInt(hex.slice(4, 6), 16) / 255;
		return 0.2126 * r + 0.7152 * g + 0.0722 * b < 0.88;
	}

	function iconColor(hex: string): string {
		return isUsable(hex) ? `#${hex}` : 'var(--color-text-muted)';
	}

	$: if (open && inputEl) {
		setTimeout(() => inputEl?.focus(), 50);
	}

	onMount(() => {
		function handleKey(e: KeyboardEvent) {
			if (e.key === 'Escape' && open) close();
		}
		window.addEventListener('keydown', handleKey);
		return () => window.removeEventListener('keydown', handleKey);
	});
</script>

{#if open}
	<!-- svelte-ignore a11y-no-static-element-interactions a11y-click-events-have-key-events -->
	<div class="overlay" on:click|self={close}>
		<div class="modal" role="dialog" aria-label="Pick an icon">
			<div class="modal-header">
				<span class="modal-title">Choose an icon</span>
				<button class="close-btn" on:click={close} aria-label="Close">✕</button>
			</div>

			<div class="search-row">
				<input
					bind:this={inputEl}
					bind:value={query}
					type="text"
					placeholder="Search {allIcons.length} icons…"
					class="search-input"
				/>
				{#if current}
					<button class="clear-btn" on:click={clear}>Remove icon</button>
				{/if}
			</div>

			<p class="result-count">
				{filtered.length > 120
					? `Showing 120 of ${filtered.length} — type to narrow`
					: `${filtered.length} result${filtered.length === 1 ? '' : 's'}`}
			</p>

			<div class="grid">
				{#each visible as icon (icon.slug)}
					<button
						class="icon-btn"
						class:selected={icon.slug === current}
						on:click={() => select(icon.slug)}
						title={icon.title}
					>
						<span class="icon-wrap" style="color:{iconColor(icon.hex)}">
							<svg
								role="img"
								viewBox="0 0 24 24"
								width="20"
								height="20"
								fill="currentColor"
								aria-hidden="true"
							>
								<path d={icon.path} />
							</svg>
						</span>
						<span class="icon-label">{icon.title}</span>
					</button>
				{/each}
			</div>
		</div>
	</div>
{/if}

<style>
	.overlay {
		position: fixed;
		inset: 0;
		background: rgba(28, 23, 20, 0.4);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
	}

	.modal {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		width: min(640px, 96vw);
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		box-shadow: 0 8px 32px rgba(28, 23, 20, 0.12);
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.modal-title {
		font-size: 14px;
		font-weight: 500;
		color: var(--color-text);
	}

	.close-btn {
		background: none;
		border: none;
		cursor: pointer;
		color: var(--color-text-muted);
		font-size: 16px;
		padding: 2px 6px;
		border-radius: var(--radius);
		transition: color 0.1s;
	}

	.close-btn:hover {
		color: var(--color-text);
	}

	.search-row {
		display: flex;
		gap: 8px;
		padding: 12px 20px;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.search-input {
		flex: 1;
		background: var(--color-surface-alt);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		color: var(--color-text);
		padding: 8px 12px;
		font-family: var(--font-sans);
		font-size: 14px;
		outline: none;
		transition: border-color 0.12s;
	}

	.search-input:focus {
		border-color: var(--color-accent);
	}

	.search-input::placeholder {
		color: var(--color-text-muted);
	}

	.clear-btn {
		background: none;
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		color: var(--color-text-muted);
		padding: 8px 12px;
		font-family: var(--font-sans);
		font-size: 13px;
		cursor: pointer;
		white-space: nowrap;
		transition: color 0.12s, border-color 0.12s;
	}

	.clear-btn:hover {
		color: var(--color-expense);
		border-color: var(--color-expense);
	}

	.result-count {
		padding: 6px 20px;
		font-size: 11px;
		color: var(--color-text-muted);
		letter-spacing: 0.04em;
		flex-shrink: 0;
		border-bottom: 1px solid var(--color-border);
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
		gap: 4px;
		padding: 12px;
		overflow-y: auto;
	}

	.icon-btn {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		padding: 10px 6px;
		background: none;
		border: 1px solid transparent;
		border-radius: var(--radius);
		cursor: pointer;
		transition:
			background 0.1s,
			border-color 0.1s;
		text-align: center;
	}

	.icon-btn:hover {
		background: var(--color-surface-alt);
		border-color: var(--color-border);
	}

	.icon-btn.selected {
		background: var(--color-accent-light);
		border-color: var(--color-accent);
	}

	.icon-wrap {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: var(--radius);
		background: var(--color-surface-alt);
		border: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.icon-btn.selected .icon-wrap {
		background: var(--color-surface);
	}

	.icon-label {
		font-size: 10px;
		color: var(--color-text-muted);
		line-height: 1.2;
		max-width: 80px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
