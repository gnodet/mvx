---
title: mvx - Universal Build Environment Bootstrap
description: Universal build environment bootstrap tool that goes beyond Maven. Zero dependencies, cross-platform, universal tools.
name: mvx
simple-name: mvx
image: mvx-logo.svg
social-github: gnodet/mvx
layout: index
---

<link rel="stylesheet" href="{site.url}css/docusaurus-style.css">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-tomorrow.min.css">

<!-- Hero Section -->
<section class="hero">
    <div class="hero-container">
        <h1 class="hero-title">mvx</h1>
        <p class="hero-subtitle">Universal Build Environment Bootstrap</p>
        <a href="{site.url}getting-started/" class="hero-cta">Get Started - 2min ⏱️</a>
    </div>
</section>

<!-- Features Section -->
<section class="features">
    <div class="features-container">
        <div class="features-grid">
            <div class="feature-card">
                <div class="feature-icon">🚀</div>
                <h3 class="feature-title">Zero Dependencies</h3>
                <p class="feature-description">
                    No need to install Java, Maven, Node.js, or Go. mvx <strong>downloads and manages everything</strong>
                    automatically. New team members can start building immediately with a single command.
                </p>
            </div>

   <div class="feature-card">
       <div class="feature-icon">🌍</div>
       <h3 class="feature-title">Cross-Platform Universal</h3>
       <p class="feature-description">
           Works seamlessly on <strong>Linux, macOS, and Windows</strong>. ARM64 and x86_64 architectures
           supported. One tool, any platform, consistent results everywhere.
       </p>
   </div>

   <div class="feature-card">
       <div class="feature-icon">🔧</div>
       <h3 class="feature-title">Universal Tool Management</h3>
       <p class="feature-description">
           Supports <strong>Maven, Go, Node.js, Java</strong>, and more. One configuration file to manage
           your entire development environment. Version-specific tool isolation per project.
       </p>
   </div>

   <div class="feature-card">
       <div class="feature-icon">⚡</div>
       <h3 class="feature-title">Simple Configuration</h3>
       <p class="feature-description">
           Define your tools and commands in a simple <strong>JSON5 configuration file</strong>.
           Custom commands become top-level commands. Command hooks and overrides for maximum flexibility.
       </p>
   </div>
</div>
</div>
</section>

<!-- Quick Start Section -->
<section class="quick-start">
    <div class="quick-start-container">
        <h2 class="quick-start-title">Get Started in Seconds</h2>
        <div class="quick-start-grid">
            <div class="quick-start-step">
                <div class="step-number">1</div>
                <h3>Install</h3>
                <pre><code class="language-bash">curl -fsSL https://raw.githubusercontent.com/gnodet/mvx/main/install-mvx.sh | sh</code></pre>
            </div>
            <div class="quick-start-step">
                <div class="step-number">2</div>
                <h3>Initialize</h3>
                <pre><code class="language-bash">mvx init</code></pre>
            </div>
            <div class="quick-start-step">
                <div class="step-number">3</div>
                <h3>Configure</h3>
                <pre><code class="language-json">{
  "tools": {
    "maven": { "version": "3.9.6" },
    "java": { "version": "21" }
  },
  "commands": {
    "build": { "script": "mvn clean install" }
  }
}</code></pre>
            </div>
            <div class="quick-start-step">
                <div class="step-number">4</div>
                <h3>Build</h3>
                <pre><code class="language-bash">./mvx setup && ./mvx build</code></pre>
            </div>
        </div>
    </div>
</section>

<!-- Use Cases Section -->
<section class="use-cases">
    <div class="use-cases-container">
        <h2 class="use-cases-title">Perfect For</h2>
        <div class="use-cases-grid">
            <div class="use-case-card">
                <h3>🏢 Enterprise Teams</h3>
                <p>Eliminate "works on my machine" problems. Consistent development environments across all team members.</p>
            </div>
            <div class="use-case-card">
                <h3>🔄 CI/CD Pipelines</h3>
                <p>Reproducible builds in any environment. No more dependency installation scripts in your CI.</p>
            </div>
            <div class="use-case-card">
                <h3>📚 Open Source Projects</h3>
                <p>Lower the barrier for contributors. One command setup for any technology stack.</p>
            </div>
            <div class="use-case-card">
                <h3>🎓 Educational Projects</h3>
                <p>Students can focus on learning, not environment setup. Works on any school computer.</p>
            </div>
        </div>
    </div>
</section>