"use client";

import { useState, useEffect, useCallback, useRef, forwardRef } from "react";
import { Search, X, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface SearchInputProps {
  value?: string;
  onChange?: (value: string) => void;
  onSearch?: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
  isLoading?: boolean;
  showClearButton?: boolean;
  autoFocus?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
  inputClassName?: string;
  disabled?: boolean;
}

export const SearchInput = forwardRef<HTMLInputElement, SearchInputProps>(
  (
    {
      value: controlledValue,
      onChange,
      onSearch,
      placeholder = "Search...",
      debounceMs = 300,
      isLoading = false,
      showClearButton = true,
      autoFocus = false,
      size = "md",
      className,
      inputClassName,
      disabled = false,
    },
    ref
  ) => {
    const [internalValue, setInternalValue] = useState(controlledValue ?? "");
    const debounceRef = useRef<NodeJS.Timeout | null>(null);
    const isControlled = controlledValue !== undefined;

    const value = isControlled ? controlledValue : internalValue;

    useEffect(() => {
      if (isControlled) {
        setInternalValue(controlledValue);
      }
    }, [controlledValue, isControlled]);

    const debouncedSearch = useCallback(
      (searchValue: string) => {
        if (debounceRef.current) {
          clearTimeout(debounceRef.current);
        }

        debounceRef.current = setTimeout(() => {
          onSearch?.(searchValue);
        }, debounceMs);
      },
      [debounceMs, onSearch]
    );

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.target.value;

      if (!isControlled) {
        setInternalValue(newValue);
      }

      onChange?.(newValue);
      debouncedSearch(newValue);
    };

    const handleClear = () => {
      if (!isControlled) {
        setInternalValue("");
      }

      onChange?.("");
      onSearch?.("");

      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") {
        if (debounceRef.current) {
          clearTimeout(debounceRef.current);
        }
        onSearch?.(value);
      }

      if (e.key === "Escape") {
        handleClear();
      }
    };

    useEffect(() => {
      return () => {
        if (debounceRef.current) {
          clearTimeout(debounceRef.current);
        }
      };
    }, []);

    const sizeClasses = {
      sm: {
        container: "h-8",
        input: "text-sm pl-8 pr-8",
        icon: "h-4 w-4 left-2",
        clear: "h-4 w-4 right-2",
      },
      md: {
        container: "h-10",
        input: "text-sm pl-10 pr-10",
        icon: "h-4 w-4 left-3",
        clear: "h-4 w-4 right-3",
      },
      lg: {
        container: "h-12",
        input: "text-base pl-12 pr-12",
        icon: "h-5 w-5 left-4",
        clear: "h-5 w-5 right-4",
      },
    };

    const sizes = sizeClasses[size];

    return (
      <div className={cn("relative", sizes.container, className)}>
        <div
          className={cn(
            "absolute top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none",
            sizes.icon
          )}
        >
          {isLoading ? (
            <Loader2 className="animate-spin h-full w-full" />
          ) : (
            <Search className="h-full w-full" />
          )}
        </div>

        <input
          ref={ref}
          type="text"
          value={value}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          autoFocus={autoFocus}
          disabled={disabled}
          className={cn(
            "w-full h-full rounded-lg border border-slate-200 bg-white",
            "placeholder:text-slate-400 text-slate-900",
            "focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500",
            "disabled:opacity-50 disabled:cursor-not-allowed",
            sizes.input,
            inputClassName
          )}
        />

        {showClearButton && value && !disabled && (
          <button
            type="button"
            onClick={handleClear}
            className={cn(
              "absolute top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600",
              "p-0.5 rounded hover:bg-slate-100",
              sizes.clear
            )}
          >
            <X className="h-full w-full" />
          </button>
        )}
      </div>
    );
  }
);

SearchInput.displayName = "SearchInput";

// Custom hook for debounced search
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}

// Search with suggestions (combobox style)
interface SearchWithSuggestionsProps<T> extends Omit<SearchInputProps, "onSearch"> {
  suggestions: T[];
  renderSuggestion: (item: T, index: number) => React.ReactNode;
  onSelect: (item: T) => void;
  getSuggestionKey: (item: T) => string;
  showNoResults?: boolean;
  noResultsText?: string;
}

export function SearchWithSuggestions<T>({
  suggestions,
  renderSuggestion,
  onSelect,
  getSuggestionKey,
  showNoResults = true,
  noResultsText = "No results found",
  value,
  onChange,
  ...props
}: SearchWithSuggestionsProps<T>) {
  const [isOpen, setIsOpen] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!isOpen) return;

    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev < suggestions.length - 1 ? prev + 1 : prev
        );
        break;
      case "ArrowUp":
        e.preventDefault();
        setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : prev));
        break;
      case "Enter":
        e.preventDefault();
        if (highlightedIndex >= 0 && suggestions[highlightedIndex]) {
          onSelect(suggestions[highlightedIndex]);
          setIsOpen(false);
        }
        break;
      case "Escape":
        setIsOpen(false);
        break;
    }
  };

  const showDropdown = isOpen && (suggestions.length > 0 || (showNoResults && value));

  return (
    <div ref={containerRef} className="relative">
      <SearchInput
        value={value}
        onChange={(v) => {
          onChange?.(v);
          setIsOpen(true);
          setHighlightedIndex(-1);
        }}
        onSearch={() => setIsOpen(true)}
        {...props}
      />

      {showDropdown && (
        <div
          className="absolute z-50 w-full mt-1 bg-white border border-slate-200 rounded-lg shadow-lg overflow-hidden"
          onKeyDown={handleKeyDown}
        >
          {suggestions.length > 0 ? (
            <div className="max-h-64 overflow-y-auto">
              {suggestions.map((item, index) => (
                <button
                  key={getSuggestionKey(item)}
                  onClick={() => {
                    onSelect(item);
                    setIsOpen(false);
                  }}
                  className={cn(
                    "w-full text-left px-3 py-2 hover:bg-slate-50",
                    highlightedIndex === index && "bg-slate-100"
                  )}
                >
                  {renderSuggestion(item, index)}
                </button>
              ))}
            </div>
          ) : showNoResults && value ? (
            <div className="px-3 py-4 text-center text-slate-500 text-sm">
              {noResultsText}
            </div>
          ) : null}
        </div>
      )}
    </div>
  );
}
