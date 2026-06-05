# SVG Quick Reference Guide 🎯

## 🎨 Enhanced Icons at a Glance

### Header Icons
| Icon | Animation | Duration | Effect |
|------|-----------|----------|--------|
| **Logo** | Rotating diamond + gradient cycle | 20s / 3s | Glow filter + shine sweep |
| **Queue Indicator** | Stroke-dash animation | 2s | Gradient stroke |

### Navigation Pills
| Icon | Animation | Duration | Effect |
|------|-----------|----------|--------|
| **Search** 🔍 | Handle dash reveal | 1.5s | Scale on hover |
| **Queue** 📋 | Staggered opacity pulse | 2s | 0.3s delays |
| **Mirrors** 🌍 | Radius pulse + ripple | 2s | Center dot pulse |
| **About** ℹ️ | Expanding wave | 2s | Fade synchronization |

### Action Buttons
| Icon | Animation | Duration | Effect |
|------|-----------|----------|--------|
| **Download** ⬇️ | Vertical gradient + dash | 1.2s | Scale 1.2x + green glow |
| **Install** 📦 | Same + box animation | 1s | Scale 1.2x + blue glow |
| **Select All** ✓ | Checkmark dash reveal | 1.5s | Scale 1.2x hover |
| **Clear** ✕ | Circle backdrop | - | Scale 1.2x hover |

### Queue Actions
| Icon | Animation | Duration | Effect |
|------|-----------|----------|--------|
| **Refresh** 🔄 | Rotating circle | 3s | Gradient stroke |
| **Delete** 🗑️ | Vertical gradient | - | Red glow on hover |

### Empty States
| Icon | Animation | Duration | Effect |
|------|-----------|----------|--------|
| **Search Empty** | Triple-phase stroke + pulse | 3s / 2s | Inner circle expansion |
| **Queue Empty** | Staggered rectangle pulses | 2s | 0.4s cascade |
| **Mirrors Empty** | Multi-layer ripple | 2s / 3s | Expanding + fading |

### About Page
| Icon | Animation | Duration | Effect |
|------|-----------|----------|--------|
| **Large Logo** | 4-color gradient + rotation | 5s / 20s | Multiple pulses + shine |

---

## ⚡ Gradient Reference

### Common Gradient IDs
```html
<!-- Logo gradients -->
#logo-gradient        /* 3-color animated (cyan→blue→green) */
#logo-shine          /* Horizontal sweep effect */

<!-- Navigation gradients -->
#search-grad         /* Current→Transparent */
#queue-gradient      /* Current 0.8→0.4 opacity */
#mirror-grad         /* Current→Faded */

<!-- Button gradients -->
#search-btn-grad     /* White→White 0.7 */
#mirror-btn-grad     /* White→White 0.6 */
#download-grad       /* Current 0.4→1.0 */
#install-grad        /* Current 0.4→1.0 */

<!-- Empty state gradients -->
#empty-grad          /* Current 0.3→0.1 */
#queue-empty-grad    /* Current 0.3→0.1 */
#mirror-empty-grad   /* Current 0.3→0.1 */

<!-- Action gradients -->
#refresh-grad        /* Current→0.5 opacity */
#delete-grad         /* Current→0.6 opacity */

<!-- About page -->
#about-gradient      /* 4-color animated */
#about-shine         /* Horizontal sweep */
```

---

## 🔧 Filter Reference

### Common Filter IDs
```html
#logo-glow          /* feGaussianBlur(stdDeviation=2) */
#icon-glow          /* feGaussianBlur(stdDeviation=1) */
#empty-glow         /* feGaussianBlur(stdDeviation=2) */
#about-glow         /* feGaussianBlur(stdDeviation=4) */
```

**Usage**:
```html
<path filter="url(#icon-glow)" .../>
```

---

## 🎬 Animation Patterns

### Pattern 1: Stroke-Dash Reveal
```html
<path d="...">
  <animate attributeName="stroke-dasharray" 
           values="0,50; 50,0" 
           dur="1.5s" 
           repeatCount="indefinite"/>
</path>
```
**Use For**: Actions, progress indicators

### Pattern 2: Radius Pulse
```html
<circle cx="10" cy="10" r="5">
  <animate attributeName="r" 
           values="5; 7; 5" 
           dur="2s" 
           repeatCount="indefinite"/>
</circle>
```
**Use For**: Attention, loading states

### Pattern 3: Opacity Fade
```html
<circle opacity="1">
  <animate attributeName="opacity" 
           values="1; 0.5; 1" 
           dur="2s" 
           repeatCount="indefinite"/>
</circle>
```
**Use For**: Breathing effect, passive elements

### Pattern 4: Rotation
```html
<path>
  <animateTransform attributeName="transform" 
                    type="rotate" 
                    from="0 10 10" 
                    to="360 10 10" 
                    dur="3s" 
                    repeatCount="indefinite"/>
</path>
```
**Use For**: Continuous activity, loading

### Pattern 5: Color Cycle
```html
<stop offset="0%" stop-color="#00d4ff">
  <animate attributeName="stop-color" 
           values="#00d4ff; #0099ff; #00d4ff" 
           dur="3s" 
           repeatCount="indefinite"/>
</stop>
```
**Use For**: Ambient background motion

---

## 🎨 CSS Classes Reference

### Icon Classes
```css
.animated-logo      /* Header logo with hover glow */
.nav-icon          /* Navigation pill icons */
.button-icon       /* Primary button icons */
.action-icon       /* Icon button actions */
.empty-icon        /* Empty state illustrations */
.about-logo        /* About page large logo */
.queue-icon        /* Queue indicator icon */
```

### Hover Effects
```css
/* Scale transforms */
.logo-icon:hover .animated-logo         { scale: 1.0 → glow: 8→16px }
.nav-pill:hover .nav-icon              { scale: 1.1 }
.primary-button:hover .button-icon     { scale: 1.15 + rotate: 5deg }
.icon-button:hover .action-icon        { scale: 1.2 }

/* Contextual glows */
.icon-button.success:hover             { drop-shadow: 0 0 12px green }
.icon-button.primary:hover             { drop-shadow: 0 0 12px blue }
.icon-button.danger:hover              { drop-shadow: 0 0 12px red }
```

### Active States
```css
.primary-button:active .button-icon    { scale: 0.95 }
.icon-button:active .action-icon       { scale: 0.9 }
.nav-pill.active .nav-icon             { filter: drop-shadow(0 0 6px) }
```

---

## ⏱️ Timing Reference

### Duration Guidelines
| Speed | Duration | Use Case |
|-------|----------|----------|
| **Instant** | 0s | Disabled animations (reduced motion) |
| **Fast** | 0.3s | Hover/active feedback |
| **Medium** | 1-2s | Attention-grabbing animations |
| **Slow** | 2-3s | Calming/ambient effects |
| **Ambient** | 5-20s | Background atmosphere |

### Easing Functions
| Function | Value | Use Case |
|----------|-------|----------|
| **Elastic** | `cubic-bezier(0.34, 1.56, 0.64, 1)` | Hover transforms |
| **Ease-in-out** | `ease-in-out` | Opacity fades |
| **Linear** | `linear` | Rotations, constant motion |

---

## 🎨 Color Values

### Primary Palette
```css
--cyan:     #00d4ff;  /* Primary tech color */
--blue:     #0099ff;  /* Secondary, trust */
--green:    #00ff88;  /* Success, accent */
--magenta:  #ff00ff;  /* Special (about page) */
```

### Opacity Levels
```css
--fill-subtle:      0.05;  /* Background fills */
--fill-light:       0.1;   /* Pulsing elements */
--stroke-faded:     0.2;   /* Empty states */
--stroke-light:     0.4;   /* Gradient ends */
--stroke-medium:    0.6;   /* Standard elements */
--stroke-strong:    0.8;   /* Gradient starts */
--fill-primary:     0.95;  /* Icon fills */
```

---

## 🔍 Quick Search

### Find by Effect

**Need Glow?**
- Use: `filter="url(#icon-glow)"`
- IDs: `logo-glow`, `icon-glow`, `empty-glow`, `about-glow`

**Need Gradient?**
- Use: `stroke="url(#gradient-id)"` or `fill="url(#gradient-id)"`
- Common: `logo-gradient`, `search-grad`, `empty-grad`

**Need Animation?**
- Stroke-dash: `<animate attributeName="stroke-dasharray" .../>`
- Rotation: `<animateTransform type="rotate" .../>`
- Pulse: `<animate attributeName="r" .../>`
- Fade: `<animate attributeName="opacity" .../>`

**Need Hover Effect?**
- Add class: `.action-icon`, `.button-icon`, `.nav-icon`
- Automatic: scale + glow on hover

---

## 📋 Common Tasks

### Add New Icon with Glow
```html
<svg class="action-icon">
  <defs>
    <filter id="my-glow">
      <feGaussianBlur stdDeviation="1" result="coloredBlur"/>
      <feMerge>
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>
  <path d="..." filter="url(#my-glow)"/>
</svg>
```

### Add Gradient Stroke
```html
<svg>
  <defs>
    <linearGradient id="my-grad" x1="0" y1="0" x2="20" y2="20">
      <stop offset="0%" stop-color="currentColor"/>
      <stop offset="100%" stop-color="currentColor" stop-opacity="0.5"/>
    </linearGradient>
  </defs>
  <path d="..." stroke="url(#my-grad)" stroke-width="2.5"/>
</svg>
```

### Add Pulsing Animation
```html
<circle cx="10" cy="10" r="5" fill="currentColor">
  <animate attributeName="r" 
           values="5; 7; 5" 
           dur="2s" 
           repeatCount="indefinite"/>
  <animate attributeName="opacity" 
           values="1; 0.6; 1" 
           dur="2s" 
           repeatCount="indefinite"/>
</circle>
```

### Add Hover Glow (CSS)
```css
.my-button:hover .my-icon {
  filter: drop-shadow(0 0 12px var(--primary-color));
  transform: scale(1.2);
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}
```

---

## ⚠️ Common Issues

### Animation Not Playing
1. Check `repeatCount="indefinite"` is set
2. Verify `dur` attribute is present
3. Test in different browsers (SMIL support)
4. Check `prefers-reduced-motion` setting

### Glow Not Showing
1. Ensure filter ID is unique
2. Check `url(#filter-id)` syntax
3. Verify `<defs>` is inside `<svg>`
4. Check z-index stacking context

### Gradient Not Applying
1. Ensure gradient ID is defined in `<defs>`
2. Check `url(#gradient-id)` syntax
3. Verify `gradientUnits="userSpaceOnUse"`
4. Check stop colors are valid

### Hover Not Working
1. Add appropriate class (`.action-icon`, etc.)
2. Check CSS specificity
3. Verify `:hover` pseudo-class
4. Ensure element is not disabled

---

## 🎓 Best Practices

### DO ✅
- Use `stroke-width="2.5"` for clarity
- Add `stroke-linecap="round"` for smooth ends
- Use `repeatCount="indefinite"` for loops
- Test with `prefers-reduced-motion`
- Reuse gradient/filter IDs

### DON'T ❌
- Don't nest filters (performance)
- Don't use `dur` < 0.3s (jarring)
- Don't forget `currentColor` (theming)
- Don't over-animate (visual chaos)
- Don't ignore accessibility

---

## 🚀 Performance Tips

1. **Reuse Definitions**: One gradient, many uses
2. **GPU Acceleration**: Use CSS transforms
3. **Lazy Load**: Only animate visible elements
4. **Optimize Paths**: Round decimal values
5. **Test Performance**: Chrome DevTools > Performance

---

## 📞 Support

### Files
- `SVG_DESIGN_SYSTEM.md` - Full documentation
- `SVG_CHANGELOG.md` - All changes
- `VISUAL_SUMMARY.md` - Visual overview
- `SVG_QUICK_REFERENCE.md` - This file

### Code Locations
- HTML: `webui/index.html`
- CSS: `webui/style.css`
- Patterns: Search for `<animate`, `<filter`, `<gradient`

---

**Last Updated**: 2026-06-05  
**Version**: 1.0.0  
**Status**: Production-Ready ✅
