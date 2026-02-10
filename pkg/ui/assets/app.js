document.addEventListener('DOMContentLoaded', () => {
    const grid = document.getElementById('skills-grid');
    const searchInput = document.getElementById('search-input');
    const emptyState = document.getElementById('empty-state');

    // Modal elements
    const modal = document.getElementById('detail-modal');
    const modalTitle = document.getElementById('modal-title');
    const modalDesc = document.getElementById('modal-desc');
    const installCmd = document.getElementById('install-cmd');

    // Global state
    let allSkills = [];
    const host = window.location.host; // e.g. localhost:8080

    // Fetch Orchestration
    async function loadSkills() {
        try {
            // 1. Fetch Catalog
            const catalogRes = await fetch('/v2/_catalog');
            if (!catalogRes.ok) throw new Error(`Catalog fetch failed: ${catalogRes.status}`);
            const catalog = await catalogRes.json();
            const repos = catalog.repositories || [];

            if (repos.length === 0) {
                renderSkills([]);
                return;
            }

            // 2. Fetch Details for each Repo (Parallel)
            const skillPromises = repos.map(async (repo) => {
                try {
                    // Fetch Tags
                    const tagsRes = await fetch(`/v2/${repo}/tags/list`);
                    if (!tagsRes.ok) return null;
                    const tagsData = await tagsRes.json();
                    const tags = tagsData.tags || [];

                    if (tags.length === 0) return null;

                    // Determine "latest" or use the last tag
                    const latestTag = tags.includes('latest') ? 'latest' : tags[tags.length - 1];

                    // Fetch Manifest for metadata (using latest to get description/author)
                    const manifestRes = await fetch(`/v2/${repo}/manifests/${latestTag}`);
                    if (!manifestRes.ok) return null;
                    const manifest = await manifestRes.json();

                    const annotations = manifest.annotations || {};
                    const shortName = repo.split('/').pop();

                    return {
                        id: repo,
                        name: shortName,
                        description: annotations['com.skr.description'] || 'No description available.',
                        author: annotations['com.skr.author'] || 'Unknown Author',
                        versions: tags.map(t => ({ version: t, tag: t })), // For now version == tag
                        latestTag: latestTag
                    };

                } catch (e) {
                    console.warn(`Failed to process repo ${repo}:`, e);
                    return null;
                }
            });

            const results = await Promise.all(skillPromises);
            allSkills = results.filter(s => s !== null);

            renderSkills(allSkills);

        } catch (err) {
            console.error('Failed to load skills:', err);
            grid.innerHTML = '<p style="color:var(--md-sys-color-error); text-align:center; grid-column: 1/-1;">Error loading skills from OCI Registry.</p>';
        }
    }

    // Search
    searchInput.addEventListener('input', (e) => {
        const term = e.target.value.toLowerCase();
        const filtered = allSkills.filter(skill => {
            return (skill.name && skill.name.toLowerCase().includes(term)) ||
                (skill.description && skill.description.toLowerCase().includes(term)) ||
                (skill.author && skill.author.toLowerCase().includes(term)) ||
                (skill.id && skill.id.toLowerCase().includes(term));
        });
        renderSkills(filtered);
    });

    // Render
    function renderSkills(skills) {
        grid.innerHTML = '';

        if (skills.length === 0) {
            emptyState.classList.remove('hidden');
            return;
        } else {
            emptyState.classList.add('hidden');
        }

        skills.forEach((skill) => {
            const card = document.createElement('div');
            card.className = 'card';
            card.style.cursor = 'pointer'; // Make interactable

            // Interaction: Open Modal
            card.onclick = () => openModal(skill);

            const author = skill.author || 'Unknown Author';
            const versions = skill.versions || [];

            // Sort versions (simple alpha sort for now, ideally semver)
            versions.sort((a, b) => b.tag.localeCompare(a.tag));

            // Create 3 chip max
            const displayVersions = versions.slice(0, 3);
            let versionChips = displayVersions.map(v =>
                `<span class="chip" title="${v.tag}">${escapeHtml(v.version)}</span>`
            ).join('');

            if (versions.length > 3) {
                versionChips += `<span class="chip">+${versions.length - 3}</span>`;
            }

            // Using full repo name as ID/Title often looks better for tech users, but shortName is cleaner.
            // Let's use name (short) and put Repo ID in subhead or tooltip?
            // Current design: Title = Name, Subhead = Author.
            // Let's stick to Name.

            card.innerHTML = `
                <div class="card-content">
                    <div class="card-title">${escapeHtml(skill.name)}</div>
                    <div class="card-subhead">${escapeHtml(author)}</div>
                    <div class="card-text">${escapeHtml(skill.description)}</div>
                    <div class="chips-container">
                        ${versionChips}
                    </div>
                </div>
            `;
            grid.appendChild(card);
        });
    }

    // Modal Logic
    window.openModal = function (skill) {
        modalTitle.innerText = skill.name;
        modalDesc.innerText = skill.description;

        // Construct install command
        // Convention: host/repo:tag
        // Note: oras pull host/repo:tag
        // skr install host/repo:tag

        const cmd = `skr install ${host}/${skill.id}:${skill.latestTag}`;
        installCmd.innerText = cmd;

        modal.classList.add('open');
    }

    window.closeModal = function () {
        modal.classList.remove('open');
    }

    window.copyInstallCmd = function () {
        const text = installCmd.innerText;
        navigator.clipboard.writeText(text).then(() => {
            // Optional: visual feedback
            const original = installCmd.style.color;
            installCmd.style.color = 'var(--md-sys-color-primary)';
            setTimeout(() => installCmd.style.color = original, 200);
        });
    }

    // Close modal on click outside
    modal.addEventListener('click', (e) => {
        if (e.target === modal) closeModal();
    });

    function escapeHtml(str) {
        if (!str) return '';
        const div = document.createElement('div');
        div.innerText = str;
        return div.innerHTML;
    }

    // Initial Load
    loadSkills();
});
