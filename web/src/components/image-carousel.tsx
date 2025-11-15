import React from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "~/lib/utils";
import { Button } from "./ui/button";

// --- TYPES ---
interface ImageCarouselProps extends React.HTMLAttributes<HTMLDivElement> {
  images: { src: string; inputSrc: string }[];
}

// --- HERO SECTION COMPONENT ---
export const ImageCarousel = React.forwardRef<
  HTMLDivElement,
  ImageCarouselProps
>(({ images, className, ...props }, ref) => {
  const [currentIndex, setCurrentIndex] = React.useState(
    Math.floor(images.length / 2),
  );

  const handleNext = React.useCallback(() => {
    setCurrentIndex((prevIndex) => (prevIndex + 1) % images.length);
  }, [images.length]);

  const handlePrev = () => {
    setCurrentIndex(
      (prevIndex) => (prevIndex - 1 + images.length) % images.length,
    );
  };

  React.useEffect(() => {
    const timer = setInterval(() => {
      handleNext();
    }, 4000);
    return () => clearInterval(timer);
  }, [handleNext]);

  return (
    <div
      ref={ref}
      className={cn(
        // 'relative w-full min-h-screen flex flex-col items-center justify-center overflow-x-hidden bg-background text-foreground p-4',
        "text-foreground relative flex w-full flex-col items-center justify-center overflow-x-hidden",
        className,
      )}
      {...props}
    >
      {/* Background Gradient */}
      {/* <div className="absolute inset-0 z-0 opacity-20" aria-hidden="true">
                    <div className="absolute bottom-0 left-[-20%] right-0 top-[-10%] h-[500px] w-[500px] rounded-full bg-[radial-gradient(circle_farthest-side,rgba(128,90,213,0.3),rgba(255,255,255,0))]"></div>
                    <div className="absolute bottom-0 right-[-20%] top-[-10%] h-[500px] w-[500px] rounded-full bg-[radial-gradient(circle_farthest-side,rgba(0,123,255,0.3),rgba(255,255,255,0))]"></div>
                </div> */}

      {/* Content */}
      <div className="z-10 flex w-full flex-col items-center space-y-8 text-center md:space-y-12">
        {/* Main Showcase Section */}
        <div className="relative flex h-[450px] w-full items-center justify-center md:h-[450px]">
          {/* Carousel Wrapper */}
          <div className="relative flex h-full w-full items-center justify-center [perspective:1000px]">
            {images.map((image, index) => {
              const offset = index - currentIndex;
              const total = images.length;
              let pos = (offset + total) % total;
              if (pos > Math.floor(total / 2)) {
                pos = pos - total;
              }

              const distance = Math.abs(pos);
              const isCenter = distance === 0;
              const isVisible = distance <= 2;

              return (
                <div
                  key={index}
                  className={cn(
                    "absolute h-96 w-48 transition-all duration-500 ease-in-out md:h-[450px] md:w-64",
                    "flex items-center justify-center",
                  )}
                  style={{
                    transform: `
                                                translateX(${pos * 45}%)
                                                scale(${1 - distance * 0.1})
                                                rotateY(${pos * -10}deg)
                                            `,
                    zIndex: 10 - distance,
                    opacity: isCenter ? 1 : 1.0 - distance * 0.2,
                    filter: `
                                                blur(${distance * 2}px)
                                                brightness(${1 - distance * 0.05})
                                            `,
                    visibility: isVisible ? "visible" : "hidden",
                  }}
                >
                  <div className="relative h-full w-full">
                    {/* LAYER 1: The full-size background image (bottom). */}
                    <img
                      src={image.src}
                      // alt={image.alt}
                      // This image fills the entire container.
                      className="absolute inset-0 z-10 h-full w-full rounded-3xl object-cover"
                    />

                    {/* LAYER 2: The transparent black gradient overlay (middle). */}
                    <div
                      // This div creates the fade effect.
                      className="absolute right-0 bottom-0 left-0 z-20 h-1/2 rounded-b-3xl bg-gradient-to-t from-black/80 to-transparent"
                    />

                    {/* LAYER 3: The small image (top). */}
                    <img
                      src={image.inputSrc}
                      //   alt={image.alt}
                      className={cn(
                        // Base styles that are always applied
                        "absolute bottom-4 left-4 h-20 w-16 md:h-30 md:w-26",
                        "z-30 rounded-md object-cover",
                        "transition-all duration-500 ease-in-out", // The transition itself

                        // Conditional styles for visibility
                        isCenter
                          ? "scale-100 opacity-100" // When visible: fully opaque and normal size
                          : "pointer-events-none scale-95 opacity-0", // When hidden: transparent, slightly smaller, and not clickable
                      )}
                    />
                  </div>
                </div>
              );
            })}
          </div>

          {/* Navigation Buttons */}
          <Button
            variant="outline"
            size="icon"
            className="bg-background/50 absolute top-1/2 left-2 z-20 h-10 w-10 -translate-y-1/2 rounded-full backdrop-blur-sm sm:left-8"
            onClick={handlePrev}
          >
            <ChevronLeft className="h-5 w-5" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="bg-background/50 absolute top-1/2 right-2 z-20 h-10 w-10 -translate-y-1/2 rounded-full backdrop-blur-sm sm:right-8"
            onClick={handleNext}
          >
            <ChevronRight className="h-5 w-5" />
          </Button>
        </div>
      </div>
    </div>
  );
});
