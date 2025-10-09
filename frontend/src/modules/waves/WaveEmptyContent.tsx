import { AudioWaveform, Radio } from 'lucide-react';

export const WaveEmptyContent = () => {
	return (
		<div className="flex flex-col h-full overflow-y-auto p-4 py-0">
			<div className="flex flex-col items-center justify-center text-center rounded-2xl border border-dashed border-muted-foreground/20 bg-muted/40 p-8 h-full">
				<Radio strokeWidth={0.9} className="size-30 text-muted-foreground mb-3" />
				<h2 className="text-lg font-semibold text-foreground">No active waves right now</h2>
				<p className="mt-2 text-sm text-muted-foreground max-w-md">
					You can already start a wave by opening the <span className="font-bold">sphere</span> of any coin and launching its audio channel.
				</p>

				<p className="text-sm text-muted-foreground max-w-md mt-4 flex-row flex items-center justify-center">
					<AudioWaveform className="size-4 mr-2" />
					<span className="text-left">Soon you’ll also be able to browse and join trending waves across the entire platform.</span>
				</p>
			</div>
		</div>
	);
};
