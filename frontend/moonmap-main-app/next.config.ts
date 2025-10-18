import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
	/* config options here */
	output: 'standalone',
	// domains: [
	// 	'cdn.memeapi.io',
	// 	'coin-images.coingecko.com',
	// 	'github.com',
	// 	'avatars.moonmap.io',
	// 	'ipfs.io',
	// 	'images.pump.fun',
	// 	'pbs.twimg.com',
	// 	'thumbnails.padre.gg',
	// ],
	devIndicators: false,
	images: {
		path: '/',
		loader: 'custom',
		loaderFile: './src/modules/server/mint-image-loader.ts',
		remotePatterns: [
			{
				protocol: 'https',
				hostname: '**', // accept all hosts https
			},
			{
				protocol: 'http',
				hostname: '**', // opcional
			},
		],
	},
};

export default nextConfig;
