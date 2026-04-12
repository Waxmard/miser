<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type Category, type MerchantIcon as MerchantIconRecord } from '$lib/api';
	import IconPicker from '$lib/IconPicker.svelte';
	import MerchantIconDisplay from '$lib/MerchantIcon.svelte';
	import * as si from 'simple-icons';
	import type { SimpleIcon } from 'simple-icons';
	import { allEmoji, parseIconSlug } from '$lib/icons';

	let categories: Category[] = [];
	let merchantIcons: MerchantIconRecord[] = [];
	let loading = true;
	let error = '';

	// Icon picker state
	let pickerOpen = false;
	let pickerTarget: { type: 'category'; id: string } | { type: 'merchant'; name: string } | null =
		null;
	let pickerCurrent: string | null = null;

	// New merchant form
	let newMerchantName = '';
	let newMerchantPickerOpen = false;
	let newMerchantSlug: string | null = null;

	const bySlug = new Map<string, SimpleIcon>(Object.values(si).map((icon) => [icon.slug, icon]));
	const emojiByChar = new Map(allEmoji.map((e) => [e.emoji, e.name]));

	onMount(async () => {
		try {
			[categories, merchantIcons] = await Promise.all([
				api.categories(),
				api.merchantIcons()
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load';
		} finally {
			loading = false;
		}
	});

	function openPicker(
		target: { type: 'category'; id: string } | { type: 'merchant'; name: string },
		current: string | null
	) {
		pickerTarget = target;
		pickerCurrent = current;
		pickerOpen = true;
	}

	async function handlePickerSelect(e: CustomEvent<string | null>) {
		const slug = e.detail;
		const target = pickerTarget;
		if (!target) return;

		try {
			if (target.type === 'category') {
				const catId = target.id;
				await api.updateCategoryIcon(catId, slug);
				categories = categories.map((c) =>
					c.id === catId ? { ...c, icon: slug ?? undefined } : c
				);
			} else {
				const merchantName = target.name;
				if (slug) {
					await api.setMerchantIcon(merchantName, slug);
					const existing = merchantIcons.findIndex((m) => m.merchant_name === merchantName);
					if (existing >= 0) {
						merchantIcons = merchantIcons.map((m) =>
							m.merchant_name === merchantName ? { ...m, icon_slug: slug } : m
						);
					} else {
						merchantIcons = [
							...merchantIcons,
							{ merchant_name: merchantName, icon_slug: slug, updated_at: '' }
						];
					}
				} else {
					await api.deleteMerchantIcon(merchantName);
					merchantIcons = merchantIcons.filter((m) => m.merchant_name !== merchantName);
				}
			}
		} catch {
			// ignore — could show a toast here
		}

		pickerTarget = null;
	}

	async function handleNewMerchantPickerSelect(e: CustomEvent<string | null>) {
		newMerchantSlug = e.detail;
	}

	async function addMerchantIcon() {
		if (!newMerchantName.trim() || !newMerchantSlug) return;
		try {
			await api.setMerchantIcon(newMerchantName.trim(), newMerchantSlug);
			const existing = merchantIcons.findIndex((m) => m.merchant_name === newMerchantName.trim());
			if (existing >= 0) {
				merchantIcons = merchantIcons.map((m) =>
					m.merchant_name === newMerchantName.trim()
						? { ...m, icon_slug: newMerchantSlug! }
						: m
				);
			} else {
				merchantIcons = [
					...merchantIcons,
					{ merchant_name: newMerchantName.trim(), icon_slug: newMerchantSlug, updated_at: '' }
				];
			}
			newMerchantName = '';
			newMerchantSlug = null;
		} catch {
			// ignore
		}
	}

	async function removeMerchantIcon(name: string) {
		try {
			await api.deleteMerchantIcon(name);
			merchantIcons = merchantIcons.filter((m) => m.merchant_name !== name);
		} catch {
			// ignore
		}
	}

	type IconInfo =
		| { type: 'si'; icon: SimpleIcon }
		| { type: 'emoji'; emoji: string; name: string }
		| null;

	function iconForSlug(slug: string | null | undefined): IconInfo {
		if (!slug) return null;
		const { library, name } = parseIconSlug(slug);
		if (library === 'emoji') {
			return { type: 'emoji', emoji: name, name: emojiByChar.get(name) ?? '' };
		}
		const icon = bySlug.get(name);
		return icon ? { type: 'si', icon } : null;
	}

	function isUsable(hex: string): boolean {
		const r = parseInt(hex.slice(0, 2), 16) / 255;
		const g = parseInt(hex.slice(2, 4), 16) / 255;
		const b = parseInt(hex.slice(4, 6), 16) / 255;
		return 0.2126 * r + 0.7152 * g + 0.0722 * b < 0.88;
	}
</script>

<svelte:head>
	<title>Settings — miser</title>
</svelte:head>

<div class="page">
	<h1>Settings</h1>

	{#if loading}
		<p class="state">Loading…</p>
	{:else if error}
		<p class="state error">{error}</p>
	{:else}
		<!-- Categories -->
		<section class="section">
			<h2>Category Icons</h2>
			<p class="section-desc">
				Assign an icon to each spending category. Icons appear in the dashboard and trends view.
			</p>
			<ul class="icon-list">
				{#each categories as cat}
					{@const info = iconForSlug(cat.icon)}
					<li class="icon-row">
						<div class="icon-preview">
							{#if info?.type === 'si'}
								<span
									class="preview-wrap"
									style="color:{isUsable(info.icon.hex) ? `#${info.icon.hex}` : 'var(--color-text-muted)'}"
								>
									<svg role="img" viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
										<path d={info.icon.path} />
									</svg>
								</span>
							{:else if info?.type === 'emoji'}
								<span class="preview-wrap preview-emoji" aria-hidden="true">{info.emoji}</span>
							{:else}
								<span class="preview-empty">—</span>
							{/if}
						</div>
						<span class="icon-row-name">{cat.name}</span>
						{#if cat.icon}
							<span class="slug-badge">{cat.icon}</span>
						{/if}
						<button
							class="edit-btn"
							on:click={() => openPicker({ type: 'category', id: cat.id }, cat.icon ?? null)}
						>
							{cat.icon ? 'Change' : 'Set icon'}
						</button>
					</li>
				{/each}
			</ul>
		</section>

		<!-- Merchant icon overrides -->
		<section class="section">
			<h2>Merchant Icon Overrides</h2>
			<p class="section-desc">
				Override the icon shown for a specific merchant name. These take priority over built-in
				brand detection.
			</p>

			{#if merchantIcons.length > 0}
				<ul class="icon-list">
					{#each merchantIcons as mi}
						{@const info = iconForSlug(mi.icon_slug)}
						<li class="icon-row">
							<div class="icon-preview">
								{#if info?.type === 'si'}
									<span
										class="preview-wrap"
										style="color:{isUsable(info.icon.hex)
											? `#${info.icon.hex}`
											: 'var(--color-text-muted)'}"
									>
										<svg
											role="img"
											viewBox="0 0 24 24"
											width="18"
											height="18"
											fill="currentColor"
										>
											<path d={info.icon.path} />
										</svg>
									</span>
								{:else if info?.type === 'emoji'}
									<span class="preview-wrap preview-emoji" aria-hidden="true">{info.emoji}</span>
								{/if}
							</div>
							<span class="icon-row-name">{mi.merchant_name}</span>
							<span class="slug-badge">{mi.icon_slug}</span>
							<button
								class="edit-btn"
								on:click={() =>
									openPicker({ type: 'merchant', name: mi.merchant_name }, mi.icon_slug)}
							>
								Change
							</button>
							<button class="remove-btn" on:click={() => removeMerchantIcon(mi.merchant_name)}>
								Remove
							</button>
						</li>
					{/each}
				</ul>
			{/if}

			<div class="add-merchant">
				<h3>Add override</h3>
				<div class="add-row">
					<input
						type="text"
						bind:value={newMerchantName}
						placeholder="Merchant name (e.g. Whole Foods)"
						class="merchant-input"
					/>
					<button
						class="pick-btn"
						on:click={() => (newMerchantPickerOpen = true)}
					>
						{newMerchantSlug
							? (() => {
									const ic = iconForSlug(newMerchantSlug);
									if (!ic) return newMerchantSlug;
									return ic.type === 'si' ? `${ic.icon.title} ✓` : `${ic.emoji} ${ic.name} ✓`;
								})()
							: 'Pick icon…'}
					</button>
					<button
						class="save-btn"
						disabled={!newMerchantName.trim() || !newMerchantSlug}
						on:click={addMerchantIcon}
					>
						Save
					</button>
				</div>
			</div>
		</section>
	{/if}
</div>

<!-- Icon picker for category/existing merchant -->
<IconPicker
	bind:open={pickerOpen}
	current={pickerCurrent}
	on:select={handlePickerSelect}
/>

<!-- Icon picker for new merchant -->
<IconPicker
	bind:open={newMerchantPickerOpen}
	current={newMerchantSlug}
	on:select={handleNewMerchantPickerSelect}
/>

<!-- Preview merchant icons in the "add" row -->
{#if newMerchantName.trim() && newMerchantSlug}
	<div style="display:none">
		<MerchantIconDisplay merchant={newMerchantName} iconSlug={newMerchantSlug} />
	</div>
{/if}

<style>
	.page {
		max-width: 800px;
	}

	h1 {
		font-family: var(--font-display);
		font-size: 36px;
		font-weight: 600;
		color: var(--color-text);
		margin-bottom: 40px;
		letter-spacing: -0.3px;
	}

	/* ── Sections ─────────────────────────────────────── */
	.section {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 24px;
		margin-bottom: 24px;
	}

	h2 {
		font-size: 11px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: var(--color-text-muted);
		margin-bottom: 6px;
	}

	h3 {
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text-muted);
		margin-bottom: 10px;
		margin-top: 24px;
	}

	.section-desc {
		font-size: 13px;
		color: var(--color-text-muted);
		margin-bottom: 20px;
		line-height: 1.5;
	}

	/* ── Icon list rows ───────────────────────────────── */
	.icon-list {
		list-style: none;
		display: flex;
		flex-direction: column;
	}

	.icon-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--color-border);
	}

	.icon-row:last-child {
		border-bottom: none;
	}

	.icon-preview {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--color-surface-alt);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		flex-shrink: 0;
	}

	.preview-wrap {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.preview-emoji {
		font-size: 18px;
		line-height: 1;
	}

	.preview-empty {
		color: var(--color-text-muted);
		font-size: 14px;
	}

	.icon-row-name {
		flex: 1;
		font-size: 14px;
		font-weight: 500;
		color: var(--color-text);
	}

	.slug-badge {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
		background: var(--color-surface-alt);
		padding: 2px 7px;
		border-radius: 3px;
		border: 1px solid var(--color-border);
	}

	/* ── Buttons ──────────────────────────────────────── */
	.edit-btn,
	.pick-btn {
		background: none;
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		color: var(--color-text-muted);
		padding: 5px 12px;
		font-family: var(--font-sans);
		font-size: 12px;
		cursor: pointer;
		transition:
			color 0.12s,
			border-color 0.12s;
		white-space: nowrap;
	}

	.edit-btn:hover,
	.pick-btn:hover {
		color: var(--color-accent);
		border-color: var(--color-accent);
	}

	.remove-btn {
		background: none;
		border: 1px solid transparent;
		border-radius: var(--radius);
		color: var(--color-text-muted);
		padding: 5px 10px;
		font-family: var(--font-sans);
		font-size: 12px;
		cursor: pointer;
		transition: color 0.12s;
	}

	.remove-btn:hover {
		color: var(--color-expense);
	}

	.save-btn {
		background: var(--color-accent);
		border: none;
		border-radius: var(--radius);
		color: #ffffff;
		padding: 8px 16px;
		font-family: var(--font-sans);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: opacity 0.12s;
		white-space: nowrap;
	}

	.save-btn:disabled {
		opacity: 0.35;
		cursor: default;
	}

	.save-btn:not(:disabled):hover {
		opacity: 0.88;
	}

	/* ── Add merchant form ────────────────────────────── */
	.add-row {
		display: flex;
		gap: 8px;
		align-items: center;
		flex-wrap: wrap;
	}

	.merchant-input {
		flex: 1;
		min-width: 200px;
		background: var(--color-surface-alt);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		color: var(--color-text);
		padding: 8px 12px;
		font-family: var(--font-sans);
		font-size: 13px;
		outline: none;
		transition: border-color 0.12s;
	}

	.merchant-input:focus {
		border-color: var(--color-accent);
	}

	.merchant-input::placeholder {
		color: var(--color-text-muted);
	}

	/* ── States ───────────────────────────────────────── */
	.state {
		color: var(--color-text-muted);
	}

	.error {
		color: var(--color-expense);
	}
</style>
