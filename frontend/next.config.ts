import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
	/* config options here */
	output: 'standalone',
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
