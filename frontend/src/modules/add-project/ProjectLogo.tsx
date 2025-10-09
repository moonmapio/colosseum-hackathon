import { cn } from '@modules/utils';
import { memo, ReactEventHandler, useCallback, useState } from 'react';

type MyLogoProps = {
	preview: string;
	onLoad?: ReactEventHandler<HTMLImageElement>;
	onLoadStart?: ReactEventHandler<HTMLImageElement>;
	onLoadError?: ReactEventHandler<HTMLImageElement>;
};

export const ProjectLogo = memo(
	(props: MyLogoProps) => {
		const { preview, onLoadStart, onLoad, onLoadError } = props;
		const [hidden, setHidden] = useState(false);

		const onLoadLocalError: ReactEventHandler<HTMLImageElement> = useCallback(
			(event) => {
				onLoadError?.(event);
				setHidden(true);
			},
			[onLoadError],
		);

		const onLoadLocal: ReactEventHandler<HTMLImageElement> = useCallback(
			(event) => {
				onLoad?.(event);
				setHidden(false);
			},
			[onLoad],
		);

		return (
			// eslint-disable-next-line @next/next/no-img-element
			<img
				key={preview}
				src={preview}
				onError={onLoadLocalError}
				onLoad={onLoadLocal}
				onLoadStart={onLoadStart}
				alt="logo"
				className={cn('w-full h-full object-cover', hidden ? 'hidden' : '')}
			/>
		);
	},
	(prev, next) => prev.preview === next.preview,
);

ProjectLogo.displayName = 'ProjectLogo';
