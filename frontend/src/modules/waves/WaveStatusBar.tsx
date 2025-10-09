import { RoomAudioRenderer, RoomContext } from '@livekit/components-react';
import { useEphemeralAppStore } from '@modules/stores/main/ephemeral-hooks';
import { MintAvatar } from '@modules/trending-mints/MintAvatar';
import Waveform from '@modules/ui/waveform';
import { Room } from 'livekit-client';
import { useEffect, useState } from 'react';
import { ParticipantsMarquee } from './WaveParticipantsMarquee';
import { LiveBadge } from '@modules/ui/live-badge';
import { Button } from '@modules/ui/button';

export function WaveformIndicator({ active }: { active?: boolean }) {
	return <Waveform className={`w-5 h-5 ${active ? 'text-primary animate-pulse' : 'text-muted-foreground'}`} />;
}

export function WaveStatusBar() {
	const wave = useEphemeralAppStore((s) => s.wave);
	const resetWave = useEphemeralAppStore((s) => s.resetWave);
	const isMobile = useEphemeralAppStore((state) => state.isMobile);
	const [room] = useState(() => new Room({ adaptiveStream: true, dynacast: true }));

	useEffect(() => {
		if (!wave.joined || !wave.token) {
			return;
		}

		let mounted = true;

		const connect = async () => {
			if (mounted && wave.token) {
				await room.connect(process.env.NEXT_PUBLIC_LIVEKIT_URL!, wave.token);
				await room.localParticipant.setMicrophoneEnabled(true);
			}
		};
		connect();

		return () => {
			mounted = false;
			room.disconnect();
		};
	}, [wave.joined, wave.token, room, wave.mint]);

	if (!wave.joined || !wave.mint) return null;

	return (
		<div className="w-full flex relative px-4" id="helmer">
			<RoomContext.Provider value={room}>
				<div
					data-mobile={isMobile}
					className="left-0 right-0 bg-muted text-muted-foreground pl-2 pr-0 py-2 
                    flex items-center border border-dashed justify-between 
                    w-full h-10 rounded-lg"
				>
					{/* LEFT SIDE */}
					<div className="flex items-center gap-3 w-full ">
						{wave.uriResolved?.image && (
							<MintAvatar
								imageUrl={wave.uriResolved.image}
								height={27}
								width={27}
								alt={wave.mint}
								mint={wave.mint}
								className="border border-primary"
							/>
						)}
						<LiveBadge />
						<div className="min-w-0 flex-1">
							<ParticipantsMarquee />
							<div className="relative overflow-hidden w-full">
								<div className="marquee-content whitespace-nowrap text-sm text-primary">
									Listening <span className="font-semibold">{wave.symbol}</span> · {wave.name}
								</div>
							</div>
						</div>
						<WaveformIndicator active />
						{/* RIGHT SIDE */}
						<Button variant="default" onClick={resetWave} className="ml-2">
							Leave
						</Button>
					</div>
				</div>

				<RoomAudioRenderer />
			</RoomContext.Provider>
		</div>
	);
}
