<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';

	const nav = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/transactions', label: 'Transactions' },
		{ href: '/trends', label: 'Trends' }
	];
</script>

<div class="layout">
	<nav class="sidebar">
		<div class="logo">miser</div>
		<ul>
			{#each nav as item}
				<li>
					<a href={item.href} class:active={$page.url.pathname === item.href}>
						{item.label}
					</a>
				</li>
			{/each}
		</ul>
	</nav>
	<main class="content">
		<slot />
	</main>
</div>

<style>
	.layout {
		display: flex;
		min-height: 100vh;
	}

	.sidebar {
		width: var(--sidebar-width);
		background: var(--color-surface);
		border-right: 1px solid var(--color-border);
		padding: 24px 16px;
		display: flex;
		flex-direction: column;
		gap: 32px;
		flex-shrink: 0;
	}

	.logo {
		font-size: 20px;
		font-weight: 700;
		letter-spacing: -0.5px;
		color: var(--color-accent);
		padding: 0 8px;
	}

	ul {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	a {
		display: block;
		padding: 8px 12px;
		border-radius: var(--radius);
		color: var(--color-text-muted);
		transition: background 0.15s, color 0.15s;
		font-size: 14px;
	}

	a:hover,
	a.active {
		background: color-mix(in srgb, var(--color-accent) 12%, transparent);
		color: var(--color-text);
	}

	a.active {
		color: var(--color-accent);
	}

	.content {
		flex: 1;
		padding: 32px;
		overflow-y: auto;
	}
</style>
