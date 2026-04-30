<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';

	const nav = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/transactions', label: 'Transactions' },
		{ href: '/trends', label: 'Trends' }
	];

	let menuOpen = false;
	let innerWidth = 0;

	$: if (innerWidth > 768) menuOpen = false;
</script>

<svelte:window bind:innerWidth />

<div class="layout">
	<nav class="sidebar">
		<div class="sidebar-header">
			<div class="logo-block">
				<img src="/logo-light-wordmark.png" alt="miser" class="logo-img" />
				<div class="tagline">personal finance</div>
			</div>
			<button class="hamburger" on:click={() => (menuOpen = !menuOpen)} aria-label="Toggle menu">
				<span></span><span></span><span></span>
			</button>
		</div>
		<ul class="nav-list" class:visible={menuOpen}>
			{#each nav as item}
				<li>
					<a
						href={item.href}
						class:active={$page.url.pathname === item.href}
						on:click={() => (menuOpen = false)}
					>
						{item.label}
					</a>
				</li>
			{/each}
		</ul>
	</nav>
	<main class="content">
		<div class="content-inner">
			<slot />
		</div>
	</main>
</div>

<style>
	.layout {
		display: flex;
		min-height: 100vh;
	}

	/* ── Sidebar (desktop) ─────────────────────────────── */
	.sidebar {
		width: var(--sidebar-width);
		background: var(--color-surface);
		border-right: 1px solid var(--color-border);
		padding: 32px 24px;
		display: flex;
		flex-direction: column;
		gap: 40px;
		flex-shrink: 0;
		position: sticky;
		top: 0;
		height: 100vh;
		overflow-y: auto;
	}

	.sidebar-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
	}

	.logo-img {
		display: block;
		height: 110px;
		width: auto;
		margin-bottom: 6px;
	}

	.tagline {
		font-size: 11px;
		font-weight: 400;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.12em;
		margin-top: 5px;
	}

	.hamburger {
		display: none;
		flex-direction: column;
		gap: 5px;
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
	}

	.hamburger span {
		display: block;
		width: 22px;
		height: 1.5px;
		background: var(--color-text-muted);
	}

	/* ── Nav list ──────────────────────────────────────── */
	.nav-list {
		list-style: none;
		display: none;
		flex-direction: column;
		gap: 0;
	}

	@media (min-width: 769px) {
		.nav-list {
			display: flex;
		}
	}

	.nav-list.visible {
		display: flex;
	}

	a {
		display: block;
		padding: 9px 12px;
		color: var(--color-text-muted);
		font-size: 14px;
		font-weight: 400;
		transition: color 0.12s;
		border-left: 3px solid transparent;
	}

	a:hover {
		color: var(--color-text);
		border-left-color: var(--color-border);
	}

	a.active {
		color: var(--color-text);
		font-weight: 500;
		border-left-color: var(--color-accent);
	}

	/* ── Main content ─────────────────────────────────── */
	.content {
		flex: 1;
		padding: 40px 48px;
		overflow-y: auto;
		background: var(--color-bg);
	}

	.content-inner {
		max-width: 2200px;
		margin: 0 auto;
	}

	/* ── Mobile (≤768px) ─────────────────────────────── */
	@media (max-width: 768px) {
		.layout {
			flex-direction: column;
		}

		.sidebar {
			width: 100%;
			height: auto;
			position: relative;
			padding: 16px 20px;
			gap: 0;
			flex-direction: column;
			overflow: visible;
			border-right: none;
			border-bottom: 1px solid var(--color-border);
		}

		.sidebar-header {
			align-items: center;
		}

		.hamburger {
			display: flex;
		}

		.nav-list {
			padding-top: 12px;
		}

		a {
			padding: 10px 8px;
			border-left: none;
		}

		a:hover {
			border-left-color: transparent;
		}

		a.active {
			border-left-color: transparent;
			color: var(--color-accent);
		}

		.content {
			padding: 24px 20px;
		}
	}
</style>
