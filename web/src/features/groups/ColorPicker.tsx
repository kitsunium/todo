import { Check } from "lucide-react";
import { GROUP_COLOR_LIST } from "../../components/ui/group-color";
import { cn } from "../../lib/cn";

export function ColorPicker({ value, onChange }: { value: string; onChange: (c: string) => void }) {
  return (
    <div role="radiogroup" aria-label="Color" className="flex flex-wrap gap-2">
      {GROUP_COLOR_LIST.map((c) => {
        const on = c.name === value;
        return (
          <button
            key={c.name}
            type="button"
            role="radio"
            aria-checked={on}
            aria-label={c.name}
            onClick={() => onChange(c.name)}
            className={cn(
              "flex size-7 items-center justify-center rounded-full outline-none transition-transform duration-150 hover:scale-110",
              "focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-surface",
              on && "ring-2 ring-offset-2 ring-offset-surface",
            )}
            style={{ backgroundColor: c.hex, ...(on ? { ["--tw-ring-color" as string]: c.hex } : {}) }}
          >
            {on ? <Check className="size-3.5 text-white" strokeWidth={3} /> : null}
          </button>
        );
      })}
    </div>
  );
}
