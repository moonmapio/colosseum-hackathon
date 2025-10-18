'use client';

import { useAppStore } from '@modules/stores/main/app-hooks';
import { Input } from '@modules/ui/input';
import { cn } from '@modules/utils';
import { useDebounce } from '@uidotdev/usehooks';
import { CloudUploadIcon, ImageOff, ImagePlus, Link2 } from 'lucide-react';
import { ChangeEvent, ReactEventHandler, useCallback, useEffect, useRef, useState } from 'react';
import { useController, useFormContext } from 'react-hook-form';
import { AvatarRing } from './AvatarRing';
import { CreateProjectInput } from './DialogBody';
import { ProjectLogo } from './ProjectLogo';
import { RemoveCoinButton } from './RemoveCoinButton';
import { useUploadLogo } from './useUploadLogo';
import { Tooltip, TooltipContent, TooltipTrigger } from '@modules/ui/tooltip';

type Props = {
	className?: string;
};

const useOnUploadFile = () => {
	const uploadLogo = useUploadLogo();

	const [uploading, setUploading] = useState(false);
	const form = useFormContext<CreateProjectInput>();

	const { setError, setValue } = form;
	const onUpload = useCallback(
		async (e: ChangeEvent<HTMLInputElement>) => {
			try {
				setUploading(true);
				if (uploading) {
					return;
				}

				const file = e.target.files?.[0];
				if (!file) {
					throw new Error('No file selected');
				}

				form.clearErrors('imageUrl');
				setValue('imageUrl', '');

				uploadLogo(file)
					.then((url) => {
						if (!url) {
							throw new Error('no url found after trying the upload');
						}
						form.setValue('imageUrl', url);
					})
					.catch((err) => setError('imageUrl', { message: err.message }))
					.finally(() => setUploading(false));
			} catch (err: unknown) {
				if (err instanceof Error) {
					form.clearErrors('imageUrl');
				}

				setUploading(false);
			}
		},
		[uploading, form, setError, setValue, uploadLogo],
	);

	return { onUpload, uploading };
};

export function LogoPicker(props: Props) {
	const { className } = props;
	const form = useFormContext<CreateProjectInput>();
	const { clearErrors, setError, control, setValue } = form;

	const imageErr = form.formState.errors.imageUrl?.message;
	const { field } = useController({ control, name: 'imageUrl' });

	const urlInputRef = useRef<HTMLInputElement | null>(null);
	const fileRef = useRef<HTMLInputElement | null>(null);
	const [ringState, setRingState] = useState<'active' | 'ready' | 'failed' | ''>('');

	const debouncedUrl = useDebounce(field.value, 500);

	const { uploading, onUpload } = useOnUploadFile();
	const disableInput = useAppStore((state) => state.draftProjectLogoUrl);
	const draftprojectLogoStatus = useAppStore((state) => state.draftProjectLogo?.status);
	const pending = draftprojectLogoStatus === 'pending' || draftprojectLogoStatus === 'processing' || draftprojectLogoStatus === 'uploaded';

	useEffect(() => {
		if (!field.value) {
			setRingState('');
		}
	}, [field.value]);

	const onUrlChange = useCallback(
		(e: ChangeEvent<HTMLInputElement>) => {
			setValue('imageUrl', e.target.value);
		},
		[setValue],
	);

	const onLoadError = useCallback(() => {
		const msg = `Unsupported link as image url: ${debouncedUrl}`;
		setRingState('failed');
		setError('imageUrl', { message: msg });
	}, [debouncedUrl, setError]);

	const onLoad: ReactEventHandler<HTMLImageElement> = useCallback(() => {
		clearErrors('imageUrl');
		setRingState('ready');
	}, [clearErrors]);

	const onLoadStart: ReactEventHandler<HTMLImageElement> = useCallback(() => {
		clearErrors('imageUrl');
		setRingState('active');
	}, [clearErrors]);

	const onStartUpload = useCallback(() => {
		if (uploading || pending) {
			return;
		}
		fileRef.current?.click();
	}, [uploading, pending]);

	return (
		<div className={cn('space-y-2', className)}>
			<div className="flex flex-col items-center gap-4">
				<div className="relative">
					<div
						data-uploading={uploading || pending}
						data-error={Boolean(imageErr)}
						onClick={onStartUpload}
						className={cn(
							'relative w-28 h-28 rounded-full overflow-hidden bg-muted',
							'border grid place-items-center data-[error=true]:border-destructive data-[error=true]:border-2',
							'data-[uploading=true]:hover:cursor-not-allowed data-[uploading=false]:hover:cursor-pointer',
						)}
					>
						{!!debouncedUrl && !pending && (
							<ProjectLogo onLoadStart={onLoadStart} onLoad={onLoad} onLoadError={onLoadError} preview={debouncedUrl} />
						)}
						{!!imageErr && <ImageOff className={cn('size-12 text-muted-foreground/60', pending ? 'hidden' : '')} strokeWidth={1.3} />}
						<CloudUploadIcon className={cn('size-6 animate-bounce text-muted-foreground', pending ? '' : 'hidden')} />
						{!imageErr && !debouncedUrl && (
							<ImagePlus className={cn('size-12 text-muted-foreground/60', pending ? 'hidden' : '')} strokeWidth={1.3} />
						)}
						<AvatarRing ringState={ringState} />
					</div>
					{!!debouncedUrl && !pending && <RemoveCoinButton />}
				</div>

				<div className="grid w-full">
					<input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={onUpload} />
				</div>
			</div>

			<div>
				<label className="text-sm flex-row flex font-medium items-center mb-1">
					<div className="mr-1">
						<Link2 className="size-3.5" />
					</div>
					Image URL
				</label>
				<Tooltip>
					<TooltipTrigger>
						<Input
							ref={urlInputRef}
							onChange={onUrlChange}
							value={field.value ?? ''}
							disabled={!!disableInput}
							className="!text-[10px] bg-background h-9"
							placeholder="https://storage.moonmap.io/mycoin.png"
						/>
					</TooltipTrigger>
					{!!field.value && <TooltipContent className="w-[300px] h-auto overflow-hidden break-all">{field.value}</TooltipContent>}
				</Tooltip>
			</div>
		</div>
	);
}
