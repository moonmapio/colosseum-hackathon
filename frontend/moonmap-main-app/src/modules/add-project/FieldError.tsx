export function FieldError({ msg }: { msg?: string }) {
	if (!msg) return null;
	return <p className="text-xs text-muted-foreground mt-1">{msg}</p>;
}
