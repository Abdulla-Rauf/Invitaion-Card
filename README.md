# 💍 Muhammed Fahad & Samar Salman — Wedding Invitation

A digital wedding invitation website crafted for **Muhammed Fahad** and **Samar Salman** to celebrate their wedding reception on **Sunday, 29 November 2026** at the **GLAMORRA International Convention Centre**, Pookkottumpadam, Kerala.

Built with a **luxury + romantic + Islamic + minimal + elegant** visual aesthetic, prioritizing mobile-first guest experiences, smooth animations, and clean separation between Go backend, HTML templates, CSS, JavaScript, and JSON data.

---

## 🌟 Architecture & Directory Structure

```
wedding-invitation/
├── main.go                       # High-performance Go HTTP server & API endpoints
├── go.mod                        # Go module configuration
├── templates/
│   ├── index.html                # Main master layout
│   └── partials/
│       ├── head.html             # SEO, OpenGraph, theme colors, typography
│       ├── envelope.html         # Interactive royal wax seal & cover screen
│       ├── hero.html             # Bismillah calligraphy & Surah Ar-Rum ayah
│       ├── parents.html          # Families' blessings and invitations
│       ├── couple.html           # Groom (Muhammed Fahad) & Bride (Samar Salman)
│       ├── event.html            # Wedding reception details, time & itinerary
│       ├── countdown.html        # Live ticking countdown to 29 November 2026
│       ├── venue.html            # GLAMORRA Convention Centre, map & navigation
│       ├── dua.html              # Prophetic Sunnah blessing (Barakallahu laka...)
│       ├── rsvp.html             # Interactive RSVP attendance & prayers form
│       └── footer.html           # Host family contacts, gratitude & closure
├── static/
│   ├── css/
│   │   └── style.css             # Vanilla CSS design system (Emerald & Antique Gold)
│   ├── js/
│   │   └── app.js                # Vanilla JavaScript (Wax seal, Canvas petals, Audio, RSVP)
│   ├── images/
│   │   ├── og-preview.png        # Social sharing banner
│   │   └── hero-art.jpg          # High-resolution royal Islamic wedding backdrop
│   ├── audio/
│   │   └── wedding-melody.mp3    # Gentle ambient celebration audio
│   └── fonts/                    # Font directory (Google Fonts loaded with fallback)
├── data/
│   ├── invitation.json           # Declarative event, couple, and venue metadata
│   └── rsvps.json                # Local JSON store for guest RSVPs
└── README.md                     # Documentation
```

---

## ✨ Features & Highlights

1. **Interactive Wax Seal Opening Ceremony**:
   - Guests arrive at a royal gatefold cover featuring an authentic 3D metallic wax seal engraved with the couple's monogram `F ✦ S`.
   - Tapping the wax seal unlocks the invitation with gentle chimes, ambient music, and a celebratory burst of jasmine petals.

2. **Falling Jasmine Petals & Golden Stardust**:
   - HTML5 Canvas particle animation simulating natural Malabar jasmine petals and radiant gold dust drifting softly across the screen.

3. **Sacred Islamic Verses & Dua**:
   - Holy Quran: **Surah Ar-Rum [30:21]** in authentic Arabic calligraphy with English translation.
   - Bismillah calligraphy crest.
   - Prophetic wedding prayer: *“Barakallahu laka wa baraka 'alayka wa jama'a baynakuma fee khayr”*.

4. **Couple & Family Presentation**:
   - **Groom**: Muhammed Fahad (S/o Mr Aboobacker CK & Ms Sulaikha M)
   - **Bride**: Samar Salman (D/o Mr Salman N & Ms Naseera C)
   - Interlocking gold arabesque monogram and union symbol.

5. **Reception Details & Program Itinerary**:
   - **Date**: Sunday, 29 November 2026 (19 Jumada al-Awwal 1448 AH)
   - **Time**: 11:30 AM – 03:00 PM IST
   - Event timeline covering welcome, couple arrival, banquet feast, and family blessings.

6. **Location & Direct Navigation**:
   - **Venue**: GLAMORRA International Convention Centre, Pookkottumpadam, Kerala.
   - Embedded interactive Google Map.
   - Direct button to open directions in Google Maps.
   - One-click button to copy full address to clipboard.

7. **One-Click Calendar Integration**:
   - Direct link to download `wedding-fahad-and-samar.ics` (Apple Calendar, Outlook, Google Calendar) generated on-the-fly by the Go server.

8. **Live Countdown Timer**:
   - Real-time days, hours, minutes, and seconds ticker counting down to the reception.

9. **Interactive RSVP & WhatsApp Integration**:
   - Guest attendance toggle (*Joyfully Accept* / *Regretfully Decline*).
   - Guest count selector (1, 2, 3, 4, 5+).
   - Heartfelt prayer/wishes textarea.
   - Submits directly to the Go server (`/api/rsvp`) and persists in `data/rsvps.json`.
   - Direct WhatsApp button to forward wishes to the host families.

10. **Audio Controller**:
    - Floating gold disc & equalizer button.
    - Autoplays ambient music upon wax seal break with manual pause/play controls.
    - Integrated procedural Web Audio API synthesizer fallback.

---

## 🚀 How to Run Locally

### Prerequisites
- [Go (Golang)](https://go.dev/) 1.22 or higher.

### Quick Start

1. Open a terminal in the `wedding-invitation` folder:
   ```bash
   cd "wedding-invitation"
   ```

2. Run the Go server:
   ```bash
   go run main.go
   ```

3. Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

### Custom Port
You can specify a custom port by setting the `PORT` environment variable:
```bash
PORT=3000 go run main.go
```

---

## 📱 Mobile-First Priority
Tested and optimized for mobile screens (iPhone, Android) as well as tablets and desktops.
All touch targets adhere to accessibility standards (>= 48px), and the presentation mimics a luxury physical card.
