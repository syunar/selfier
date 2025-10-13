import React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from '~/lib/utils';
import { Button } from './ui/button';

// --- TYPES ---
interface ImageCarouselProps extends React.HTMLAttributes<HTMLDivElement> {
    images: { src: string; inputSrc: string; alt: string; }[];
}

// --- HERO SECTION COMPONENT ---
export const ImageCarousel = React.forwardRef<HTMLDivElement, ImageCarouselProps>(
    ({ images, className, ...props }, ref) => {
        const [currentIndex, setCurrentIndex] = React.useState(Math.floor(images.length / 2));

        const handleNext = React.useCallback(() => {
            setCurrentIndex((prevIndex) => (prevIndex + 1) % images.length);
        }, [images.length]);

        const handlePrev = () => {
            setCurrentIndex((prevIndex) => (prevIndex - 1 + images.length) % images.length);
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
                    'relative w-full flex flex-col items-center justify-center overflow-x-hidden text-foreground',
                    className
                )}
                {...props}
            >
                {/* Background Gradient */}
                {/* <div className="absolute inset-0 z-0 opacity-20" aria-hidden="true">
                    <div className="absolute bottom-0 left-[-20%] right-0 top-[-10%] h-[500px] w-[500px] rounded-full bg-[radial-gradient(circle_farthest-side,rgba(128,90,213,0.3),rgba(255,255,255,0))]"></div>
                    <div className="absolute bottom-0 right-[-20%] top-[-10%] h-[500px] w-[500px] rounded-full bg-[radial-gradient(circle_farthest-side,rgba(0,123,255,0.3),rgba(255,255,255,0))]"></div>
                </div> */}

                {/* Content */}
                <div className="z-10 flex w-full flex-col items-center text-center space-y-8 md:space-y-12">

                    {/* Main Showcase Section */}
                    <div className="relative w-full h-[450px] md:h-[450px] flex items-center justify-center">
                        {/* Carousel Wrapper */}
                        <div className="relative w-full h-full flex items-center justify-center [perspective:1000px]">
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
                                            'absolute w-48 h-96 md:w-64 md:h-[450px] transition-all duration-500 ease-in-out',
                                            'flex items-center justify-center'
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
                                            visibility: isVisible ? 'visible' : 'hidden',
                                        }}
                                    >
                                        <div className="relative w-full h-full">

                                            {/* LAYER 1: The full-size background image (bottom). */}
                                            <img
                                                src={image.src}
                                                alt={image.alt}
                                                // This image fills the entire container.
                                                className="absolute inset-0 object-cover w-full h-full rounded-3xl z-10"
                                            />

                                            {/* LAYER 2: The transparent black gradient overlay (middle). */}
                                            <div
                                                // This div creates the fade effect.
                                                className="absolute bottom-0 left-0 right-0 h-1/2
                                                    bg-gradient-to-t from-black/80 to-transparent
                                                    rounded-b-3xl z-20"
                                            />

                                            {/* LAYER 3: The small image (top). */}
                                            <img
                                                src={image.inputSrc}
                                                alt={image.alt}
                                                className={cn(
                                                    // Base styles that are always applied
                                                    "absolute bottom-4 left-4 w-16 h-20 md:w-20 md:h-24",
                                                    "object-cover rounded-md z-30",
                                                    "transition-all duration-500 ease-in-out", // The transition itself

                                                    // Conditional styles for visibility
                                                    isCenter
                                                        ? "opacity-100 scale-100" // When visible: fully opaque and normal size
                                                        : "opacity-0 scale-95 pointer-events-none" // When hidden: transparent, slightly smaller, and not clickable
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
                            className="absolute left-2 sm:left-8 top-1/2 -translate-y-1/2 rounded-full h-10 w-10 z-20 bg-background/50 backdrop-blur-sm"
                            onClick={handlePrev}
                        >
                            <ChevronLeft className="h-5 w-5" />
                        </Button>
                        <Button
                            variant="outline"
                            size="icon"
                            className="absolute right-2 sm:right-8 top-1/2 -translate-y-1/2 rounded-full h-10 w-10 z-20 bg-background/50 backdrop-blur-sm"
                            onClick={handleNext}
                        >
                            <ChevronRight className="h-5 w-5" />
                        </Button>
                    </div>
                </div>
            </div>
        );
    }
);
