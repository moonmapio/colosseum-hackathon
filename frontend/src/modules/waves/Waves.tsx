'use client';

import { useAppStore } from '@modules/stores/main/app-hooks';
import { AudioWaveform } from 'lucide-react';
import { WaveEmptyContent } from './WaveEmptyContent';
import { MoonMapContentLayout } from '@modules/layout/MoonMapContentLayout';

export const Waves = () => {
	const selectedMenu = useAppStore((state) => state.selectedMenu);
	const show = selectedMenu === 'WAVES';
	if (!show) return null;

	return (
		<MoonMapContentLayout>
			{/* header pequeño arriba */}
			<div className="flex items-center px-4 shrink-0">
				<AudioWaveform className="w-6 h-6 mr-2" />
				<h1 className="text-md font-medium ml-1 text-muted-foreground">Waves</h1>
			</div>

			<WaveEmptyContent />
			{/* <SphereEmptyContent /> */}
		</MoonMapContentLayout>
	);
};
