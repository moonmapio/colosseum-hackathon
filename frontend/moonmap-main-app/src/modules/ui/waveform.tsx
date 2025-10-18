'use client';
import { cn } from '@modules/utils';
import * as React from 'react';

type Props = {
	active?: boolean;
	bars?: number; // cuántas barras
	size?: number; // px
	color?: string; // svg fill/stroke
	className?: string;
	speedMs?: number; // duración de animación
};
export default function AudioWaveIcon({ active = true, bars = 5, size = 20, color = 'currentColor', className, speedMs = 900 }: Props) {
	const w = bars * 3;
	const h = size / 2; // ahora alto proporcional al size
	const items = Array.from({ length: bars });

	// definimos valores relativos (escala)
	const baseY = h * 0.3;
	const baseHeight = h * 0.4;

	return (
		<svg
			width={size}
			height={size}
			viewBox={`0 0 ${w} ${h}`}
			className={cn(className, !active ? '!text-muted-foreground' : '')}
			aria-hidden
		>
			{items.map((_, i) => {
				const x = i * 3;
				const delay = (i * speedMs) / 10;
				return (
					<rect
						key={i}
						x={x + 0.5}
						y={baseY}
						width={2}
						height={baseHeight}
						rx={1}
						fill={color}
						style={
							active
								? {
										animation: `mm-wave ${speedMs}ms ${delay}ms infinite ease-in-out`,
										transformOrigin: `${x + 1.5}px ${h / 2}px`,
									}
								: undefined
						}
					/>
				);
			})}
			<style jsx>{`
				@keyframes mm-wave {
					0%,
					100% {
						y: ${baseY};
						height: ${baseHeight};
					}
					50% {
						y: ${baseY / 2};
						height: ${h * 0.8};
					}
				}
			`}</style>
		</svg>
	);
}
