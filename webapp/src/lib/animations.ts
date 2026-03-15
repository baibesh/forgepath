"use client";

import gsap from "gsap";

export function fadeInUp(element: HTMLElement, delay = 0) {
  gsap.fromTo(
    element,
    { opacity: 0, y: 20 },
    { opacity: 1, y: 0, duration: 0.4, delay, ease: "power2.out" }
  );
}

export function staggerFadeIn(elements: HTMLElement[], stagger = 0.08) {
  gsap.fromTo(
    elements,
    { opacity: 0, y: 16 },
    { opacity: 1, y: 0, duration: 0.35, stagger, ease: "power2.out" }
  );
}

export function scaleIn(element: HTMLElement, delay = 0) {
  gsap.fromTo(
    element,
    { opacity: 0, scale: 0.9 },
    { opacity: 1, scale: 1, duration: 0.3, delay, ease: "back.out(1.4)" }
  );
}
