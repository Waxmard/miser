import { mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const simpleIconsEntry = require.resolve('simple-icons');
const simpleIconsRoot = path.dirname(simpleIconsEntry);
const iconDataPath = path.join(simpleIconsRoot, 'data', 'simple-icons.json');
const iconSvgDir = path.join(simpleIconsRoot, 'icons');
const generatedRoot = path.resolve(import.meta.dirname, '..', 'src', 'lib', 'generated');
const outputDir = path.join(generatedRoot, 'simple-icons');

const PATH_PATTERN = /<path\s+d="([^"]+)"/;

function fail(message) {
	throw new Error(`build-simple-icons-map: ${message}`);
}

async function main() {
	const raw = await readFile(iconDataPath, 'utf8');
	/** @type {{ slug?: string; hex?: string }[]} */
	const icons = JSON.parse(raw);
	await rm(outputDir, { recursive: true, force: true });
	await mkdir(outputDir, { recursive: true });

	for (const icon of icons) {
		if (!icon.slug || !icon.hex) {
			fail(`icon entry missing slug or hex: ${JSON.stringify(icon)}`);
		}

		const svgPath = path.join(iconSvgDir, `${icon.slug}.svg`);
		const svg = await readFile(svgPath, 'utf8');
		const match = svg.match(PATH_PATTERN);
		if (!match) {
			fail(`unable to extract path data from ${svgPath}`);
		}

		const outputPath = path.join(outputDir, `${icon.slug}.json`);
		const contents = JSON.stringify({
			hex: icon.hex,
			path: match[1]
		});
		await writeFile(outputPath, `${contents}\n`, 'utf8');
	}
}

main().catch((error) => {
	console.error(error instanceof Error ? error.message : error);
	process.exitCode = 1;
});
