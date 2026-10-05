/**
 * Wedding Invitation Application
 * Muhammed Fahad & Samar Salman — 29 November 2026
 * Minimal Luxury Edition
 */

(function () {
  'use strict';

  // --- DOM Elements ---
  const envelopeCover = document.getElementById('envelope-cover');
  const openInvitationBtn = document.getElementById('open-invitation-btn');
  const invitationMain = document.getElementById('invitation-main');
  const petalsCanvas = document.getElementById('petals-canvas');
  const musicToggleBtn = document.getElementById('music-toggle-btn');
  const musicIconPaused = document.getElementById('music-icon-paused');
  const musicIconPlaying = document.getElementById('music-icon-playing');
  const bgAudio = document.getElementById('bg-audio');
  const scrollTopBtn = document.getElementById('scroll-top-btn');
  const toastContainer = document.getElementById('toast-container');
  const navItems = document.querySelectorAll('.nav-item');

  let audioPlaying = false;
  let webAudioSynth = null;
  const prefersReducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');

  function isReducedMotion() {
    return prefersReducedMotionQuery ? prefersReducedMotionQuery.matches : false;
  }

  // --- Initialize Application ---
  function init() {
    setupEnvelopeOpener();
    setupCoupleIntroObserver();
    setupCanvasParticles();
    setupAudioPlayer();
    setupCountdown();
    setupScrollToTop();
    setupScrollAnimations();
  }

  // --- 1. Page 1 Opening Screen to Page 2 Smooth Transition ---
  function setupEnvelopeOpener() {
    if (!openInvitationBtn) return;

    openInvitationBtn.addEventListener('click', function () {
      // 1. Resume background music ONLY if user previously saved preference as 'playing'
      try {
        const savedMusicPref = localStorage.getItem('wedding_music_pref');
        if (savedMusicPref === 'playing') {
          startBackgroundMelody();
        }
      } catch (_) {}

      // 2. Play gentle opening chime
      playSealChime();

      // 3. Animate opening screen transition (smooth scale & fade)
      envelopeCover.classList.add('unveiled');
      document.body.classList.remove('envelope-locked');
      invitationMain.setAttribute('aria-hidden', 'false');

      const reduced = isReducedMotion();

      // 4. Trigger celebration blossom petal burst if motion allowed
      if (!reduced) {
        burstParticles();
      }

      // 5. Smoothly scroll to Page 2 (The Couple Introduction) & trigger animation
      const scrollBehavior = reduced ? 'auto' : 'smooth';
      setTimeout(() => {
        const coupleSection = document.getElementById('couple');
        if (coupleSection) {
          coupleSection.scrollIntoView({ behavior: scrollBehavior, block: 'start' });
          coupleSection.classList.add('couple-intro-animated');
        } else {
          const targetPage2 = document.getElementById('invitation-main') || document.body;
          targetPage2.scrollIntoView({ behavior: scrollBehavior, block: 'start' });
        }
        checkFadeElements();
      }, reduced ? 0 : 350);

      // 6. Complete transition and hide opening screen from accessibility tree
      setTimeout(() => {
        envelopeCover.style.display = 'none';
      }, reduced ? 50 : 900);

      showToast('✨ Welcome to our wedding celebration');
    });
  }

  // --- 2. Ambient Audio & Procedural Synthesizer Fallback ---
  function setupAudioPlayer() {
    if (!musicToggleBtn) return;

    // Ensure initial state shows paused '♫'
    setAudioState(false);

    musicToggleBtn.addEventListener('click', function (e) {
      e.stopPropagation();
      if (audioPlaying) {
        pauseAudio();
        try {
          localStorage.setItem('wedding_music_pref', 'paused');
        } catch (_) {}
        showToast('🎵 Background music paused');
      } else {
        startBackgroundMelody();
        try {
          localStorage.setItem('wedding_music_pref', 'playing');
        } catch (_) {}
        showToast('🎶 Playing wedding background melody');
      }
    });

    if (bgAudio) {
      bgAudio.addEventListener('playing', () => setAudioState(true));
      bgAudio.addEventListener('pause', () => setAudioState(false));
      bgAudio.addEventListener('ended', () => setAudioState(false));
    }
  }

  function startBackgroundMelody() {
    audioPlaying = true;
    setAudioState(true);

    if (bgAudio) {
      bgAudio.volume = 0.45;
      const playPromise = bgAudio.play();

      if (playPromise !== undefined) {
        playPromise
          .then(() => {
            setAudioState(true);
          })
          .catch((err) => {
            console.log('Audio play blocked or unavailable, using Web Audio synth fallback:', err);
            initProceduralAmbientMusic();
            setAudioState(true);
          });
      }
    } else {
      initProceduralAmbientMusic();
      setAudioState(true);
    }
  }

  function pauseAudio() {
    audioPlaying = false;
    if (bgAudio && !bgAudio.paused) {
      bgAudio.pause();
    }
    if (webAudioSynth) {
      webAudioSynth.pause();
    }
    setAudioState(false);
  }

  function setAudioState(isPlaying) {
    audioPlaying = isPlaying;
    if (!musicToggleBtn) return;

    if (isPlaying) {
      musicToggleBtn.classList.add('is-playing');
      musicToggleBtn.setAttribute('aria-label', 'Pause background music');
      musicToggleBtn.setAttribute('title', 'Pause background music');
      if (musicIconPaused) {
        musicIconPaused.classList.add('hidden');
        musicIconPaused.setAttribute('aria-hidden', 'true');
      }
      if (musicIconPlaying) {
        musicIconPlaying.classList.remove('hidden');
        musicIconPlaying.setAttribute('aria-hidden', 'false');
      }
    } else {
      musicToggleBtn.classList.remove('is-playing');
      musicToggleBtn.setAttribute('aria-label', 'Play background music');
      musicToggleBtn.setAttribute('title', 'Play background music');
      if (musicIconPaused) {
        musicIconPaused.classList.remove('hidden');
        musicIconPaused.setAttribute('aria-hidden', 'false');
      }
      if (musicIconPlaying) {
        musicIconPlaying.classList.add('hidden');
        musicIconPlaying.setAttribute('aria-hidden', 'true');
      }
    }
  }

  // Procedural Web Audio API Ambient Soundscape (Halal soothing acoustic warmth)
  function initProceduralAmbientMusic() {
    if (webAudioSynth) {
      webAudioSynth.resume();
      return;
    }

    try {
      const AudioContext = window.AudioContext || window.webkitAudioContext;
      const ctx = new AudioContext();

      // Pentatonic warm notes in Hz (soft bell / acoustic tones)
      const scale = [261.63, 293.66, 329.63, 392.00, 440.00, 523.25];
      let timer = null;

      function playPluck() {
        if (!audioPlaying) return;
        const note = scale[Math.floor(Math.random() * scale.length)];
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();

        osc.type = 'sine';
        osc.frequency.setValueAtTime(note, ctx.currentTime);

        gain.gain.setValueAtTime(0.001, ctx.currentTime);
        gain.gain.exponentialRampToValueAtTime(0.04, ctx.currentTime + 0.08);
        gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 2.8);

        osc.connect(gain);
        gain.connect(ctx.destination);

        osc.start();
        osc.stop(ctx.currentTime + 3.0);

        const nextDelay = 1400 + Math.random() * 1800;
        timer = setTimeout(playPluck, nextDelay);
      }

      webAudioSynth = {
        resume: function () {
          if (ctx.state === 'suspended') ctx.resume();
          playPluck();
        },
        pause: function () {
          if (timer) clearTimeout(timer);
        }
      };

      webAudioSynth.resume();
    } catch (e) {
      console.warn('Web Audio not supported:', e);
    }
  }

  function playSealChime() {
    try {
      const AudioCtx = window.AudioContext || window.webkitAudioContext;
      const ctx = new AudioCtx();
      const chimeFreqs = [523.25, 659.25, 783.99, 1046.50];
      chimeFreqs.forEach((freq, idx) => {
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, ctx.currentTime + idx * 0.08);
        gain.gain.setValueAtTime(0.001, ctx.currentTime + idx * 0.08);
        gain.gain.exponentialRampToValueAtTime(0.06, ctx.currentTime + idx * 0.08 + 0.04);
        gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + idx * 0.08 + 1.2);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(ctx.currentTime + idx * 0.08);
        osc.stop(ctx.currentTime + idx * 0.08 + 1.3);
      });
    } catch (err) {}
  }

  // --- 3. Ambient Canvas: Delicate Blush Petals & Soft Gold Stardust ---
  function setupCanvasParticles() {
    if (!petalsCanvas || isReducedMotion()) return;
    const ctx = petalsCanvas.getContext('2d');
    let width, height;
    let particles = [];
    const isMobile = window.innerWidth < 600;
    const particleCount = isMobile ? 16 : 28;
    let animationId = null;
    let isRunning = false;

    function resize() {
      width = petalsCanvas.width = window.innerWidth;
      height = petalsCanvas.height = window.innerHeight;
    }
    resize();

    let resizeRaf = null;
    window.addEventListener('resize', () => {
      if (resizeRaf) cancelAnimationFrame(resizeRaf);
      resizeRaf = requestAnimationFrame(resize);
    }, { passive: true });

    class Particle {
      constructor(isPetal = true) {
        this.isPetal = isPetal;
        this.reset(true);
      }

      reset(init = false) {
        this.x = Math.random() * width;
        this.y = init ? Math.random() * height : -20;
        this.size = this.isPetal ? (7 + Math.random() * 7) : (1.5 + Math.random() * 2);
        this.speedY = this.isPetal ? (0.5 + Math.random() * 0.9) : (0.3 + Math.random() * 0.6);
        this.speedX = (Math.random() - 0.5) * 0.7;
        this.rotation = Math.random() * Math.PI * 2;
        this.rotationSpeed = (Math.random() - 0.5) * 0.02;
        this.opacity = this.isPetal ? (0.35 + Math.random() * 0.35) : (0.25 + Math.random() * 0.4);
        this.sway = Math.random() * Math.PI * 2;
        this.swaySpeed = 0.015 + Math.random() * 0.02;
      }

      update() {
        this.sway += this.swaySpeed;
        this.x += this.speedX + Math.sin(this.sway) * 0.5;
        this.y += this.speedY;
        this.rotation += this.rotationSpeed;

        if (this.y > height + 20 || this.x < -20 || this.x > width + 20) {
          this.reset();
        }
      }

      draw() {
        ctx.save();
        ctx.translate(this.x, this.y);
        ctx.rotate(this.rotation);
        ctx.globalAlpha = this.opacity;

        if (this.isPetal) {
          // Delicate Magnolia / Rose Petal
          ctx.beginPath();
          ctx.moveTo(0, -this.size);
          ctx.bezierCurveTo(this.size * 0.7, -this.size * 0.5, this.size * 0.6, this.size * 0.7, 0, this.size);
          ctx.bezierCurveTo(-this.size * 0.6, this.size * 0.7, -this.size * 0.7, -this.size * 0.5, 0, -this.size);
          
          const grad = ctx.createLinearGradient(0, -this.size, 0, this.size);
          grad.addColorStop(0, '#FFFFFF');
          grad.addColorStop(0.5, '#F7D6D0');
          grad.addColorStop(1, '#E2B4BD');
          ctx.fillStyle = grad;
          ctx.fill();
        } else {
          // Soft Antique Gold dust particle
          ctx.beginPath();
          ctx.arc(0, 0, this.size, 0, Math.PI * 2);
          ctx.fillStyle = '#C5A059';
          ctx.fill();
        }
        ctx.restore();
      }
    }

    for (let i = 0; i < particleCount; i++) {
      particles.push(new Particle(i % 3 !== 0));
    }

    function step() {
      if (!isRunning) return;
      ctx.clearRect(0, 0, width, height);
      for (let p of particles) {
        p.update();
        p.draw();
      }
      animationId = requestAnimationFrame(step);
    }

    function startAnimation() {
      if (isRunning || isReducedMotion()) return;
      isRunning = true;
      animationId = requestAnimationFrame(step);
    }

    function stopAnimation() {
      isRunning = false;
      if (animationId) {
        cancelAnimationFrame(animationId);
        animationId = null;
      }
    }

    // Automatically pause render loop when tab is backgrounded
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        stopAnimation();
      } else {
        startAnimation();
      }
    }, { passive: true });

    startAnimation();

    window.addBurstPetals = function () {
      if (isReducedMotion()) return;
      const count = isMobile ? 10 : 18;
      for (let i = 0; i < count; i++) {
        const p = new Particle(true);
        p.x = width / 2 + (Math.random() - 0.5) * 140;
        p.y = height / 2 + (Math.random() - 0.5) * 140;
        p.speedY = (Math.random() - 0.5) * 3.5;
        p.speedX = (Math.random() - 0.5) * 3.5;
        particles.push(p);
      }
      setTimeout(() => {
        particles.splice(particleCount, count);
      }, 4000);
    };
  }

  function burstParticles() {
    if (window.addBurstPetals) {
      window.addBurstPetals();
    }
  }

  // --- 4. Live Accurate Countdown to 29 November 2026 (Asia/Kolkata) ---
  function setupCountdown() {
    const daysEl = document.getElementById('countdown-days');
    const hoursEl = document.getElementById('countdown-hours');
    const minsEl = document.getElementById('countdown-minutes');
    const secsEl = document.getElementById('countdown-seconds');
    const celebrationMsg = document.getElementById('countdown-celebration-msg');
    const gridEl = document.getElementById('countdown-cards-grid');

    if (!daysEl || !hoursEl || !minsEl || !secsEl) return;

    // Target wedding reception: 29 November 2026, 11:30 AM IST (Asia/Kolkata: UTC+05:30)
    const targetTimestamp = gridEl?.getAttribute('data-target') || '2026-11-29T11:30:00+05:30';
    const targetDate = new Date(targetTimestamp).getTime();
    let lastValues = { days: '', hours: '', mins: '', secs: '' };

    function updateTimer() {
      const now = new Date().getTime();
      const difference = targetDate - now;

      if (difference <= 0) {
        if (daysEl.textContent !== '00') daysEl.textContent = '00';
        if (hoursEl.textContent !== '00') hoursEl.textContent = '00';
        if (minsEl.textContent !== '00') minsEl.textContent = '00';
        if (secsEl.textContent !== '00') secsEl.textContent = '00';

        if (celebrationMsg && celebrationMsg.classList.contains('hidden')) {
          celebrationMsg.classList.remove('hidden');
        }
        return;
      }

      if (celebrationMsg && !celebrationMsg.classList.contains('hidden')) {
        celebrationMsg.classList.add('hidden');
      }

      const days = Math.floor(difference / (1000 * 60 * 60 * 24));
      const hours = Math.floor((difference % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
      const minutes = Math.floor((difference % (1000 * 60 * 60)) / (1000 * 60));
      const seconds = Math.floor((difference % (1000 * 60)) / 1000);

      const dStr = String(days).padStart(2, '0');
      const hStr = String(hours).padStart(2, '0');
      const mStr = String(minutes).padStart(2, '0');
      const sStr = String(seconds).padStart(2, '0');

      if (lastValues.days !== dStr) { daysEl.textContent = dStr; lastValues.days = dStr; }
      if (lastValues.hours !== hStr) { hoursEl.textContent = hStr; lastValues.hours = hStr; }
      if (lastValues.mins !== mStr) { minsEl.textContent = mStr; lastValues.mins = mStr; }
      if (lastValues.secs !== sStr) { secsEl.textContent = sStr; lastValues.secs = sStr; }
    }

    updateTimer();
    setInterval(updateTimer, 1000);
  }

  // --- 5. Floating Scroll-To-Top Button Handler ---
  function setupScrollToTop() {
    if (!scrollTopBtn) return;
    let scrollBtnTicking = false;
    window.addEventListener('scroll', () => {
      if (!scrollBtnTicking) {
        requestAnimationFrame(() => {
          const shouldBeVisible = window.scrollY > 350;
          if (shouldBeVisible && !scrollTopBtn.classList.contains('visible')) {
            scrollTopBtn.classList.add('visible');
          } else if (!shouldBeVisible && scrollTopBtn.classList.contains('visible')) {
            scrollTopBtn.classList.remove('visible');
          }
          scrollBtnTicking = false;
        });
        scrollBtnTicking = true;
      }
    }, { passive: true });

    scrollTopBtn.addEventListener('click', () => {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    });
  }

  // --- 7. Toast Notifications ---
  function showToast(message) {
    if (!toastContainer) return;
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.innerHTML = `<span>${message}</span>`;
    toastContainer.appendChild(toast);

    setTimeout(() => {
      if (toast.parentNode) {
        toast.parentNode.removeChild(toast);
      }
    }, 3200);
  }

  // --- 8. Page 2 Couple Introduction IntersectionObserver ---
  function setupCoupleIntroObserver() {
    const coupleSection = document.getElementById('couple');
    if (!coupleSection) return;

    if (isReducedMotion()) {
      coupleSection.classList.add('couple-intro-animated');
      return;
    }

    if ('IntersectionObserver' in window) {
      const coupleObserver = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (entry.isIntersecting) {
              entry.target.classList.add('couple-intro-animated');
              coupleObserver.unobserve(entry.target);
            }
          });
        },
        { threshold: 0.12, rootMargin: '0px 0px -20px 0px' }
      );
      coupleObserver.observe(coupleSection);
    } else {
      coupleSection.classList.add('couple-intro-animated');
    }
  }

  // --- 9. Cinematic Scroll & Section Reveal Animations ---
  function setupScrollAnimations() {
    const animSelector = '.fade-in-up, .section-header, .image-reveal, .text-reveal, .wedding-date-card-wrapper, .reception-card-wrapper, .venue-card-premium, .countdown-wrapper, .final-closing-card';
    const animElements = document.querySelectorAll(animSelector);

    if (isReducedMotion()) {
      animElements.forEach((el) => el.classList.add('in-view'));
      return;
    }

    if ('IntersectionObserver' in window) {
      const observer = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (entry.isIntersecting) {
              const delay = entry.target.getAttribute('data-delay') || 0;
              const delayMs = parseInt(delay, 10);
              if (delayMs > 0 && !isReducedMotion()) {
                setTimeout(() => {
                  entry.target.classList.add('in-view');
                }, delayMs);
              } else {
                entry.target.classList.add('in-view');
              }
              observer.unobserve(entry.target);
            }
          });
        },
        { threshold: 0.12, rootMargin: '0px 0px -30px 0px' }
      );

      animElements.forEach((el) => observer.observe(el));
    } else {
      animElements.forEach((el) => el.classList.add('in-view'));
    }

    // Dynamic listener for prefers-reduced-motion changes
    if (prefersReducedMotionQuery && prefersReducedMotionQuery.addEventListener) {
      prefersReducedMotionQuery.addEventListener('change', (e) => {
        if (e.matches) {
          document.querySelectorAll(animSelector).forEach((el) => el.classList.add('in-view'));
          const couple = document.getElementById('couple');
          if (couple) couple.classList.add('couple-intro-animated');
        }
      });
    }
  }

  function checkFadeElements() {
    const animSelector = '.fade-in-up:not(.in-view), .section-header:not(.in-view), .image-reveal:not(.in-view), .text-reveal:not(.in-view)';
    const elements = document.querySelectorAll(animSelector);
    elements.forEach((el) => {
      const rect = el.getBoundingClientRect();
      if (rect.top < window.innerHeight - 30) {
        el.classList.add('in-view');
      }
    });
  }



  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
