// ============================================================================
// MODULE D'ABONNEMENT: MODAL, FORMULAIRE, VALIDATION, SIMULATION DE PAIEMENT
// ============================================================================
// Ce script gère:
// - L'ouverture/fermeture du modal d'abonnement
// - La sélection d'un plan et le calcul du prix
// - La saisie et validation de carte (formatage simple)
// - La simulation d'un paiement (setTimeout) et persistance localStorage
// - L'état d'abonnement au chargement
// ============================================================================
// Note: C'est une simulation (aucune API réelle de paiement utilisée).
// Ne jamais stocker des numéros de carte réels côté client.
// ============================================================================
// Subscription Module
(function() {
    // Références DOM
    const subscribeBtn = document.getElementById('subscribeBtn');
    const modal = document.getElementById('subscriptionModal');
    const closeModal = document.getElementById('closeModal');
    const backToPlans = document.getElementById('backToPlans');
    const closeSuccess = document.getElementById('closeSuccess');
    const paymentButtons = document.querySelectorAll('.btn-payment');
    const paymentForm = document.getElementById('paymentForm');
    const cardForm = document.getElementById('cardForm');
    const subscriptionPlans = document.querySelector('.subscription-plans');
    const successMessage = document.getElementById('successMessage');
    const successMessage2 = document.getElementById('successMessage');
    const totalPriceEl = document.getElementById('totalPrice');
    const planNameEl = document.getElementById('planName');
    
    // État courant du modal et de la sélection
    let selectedPlan = null;
    let selectedPrice = null;
    let selectedPlanName = null;

    // Ouvrir le modal d'abonnement
    subscribeBtn?.addEventListener('click', () => {
        modal.classList.add('active');
        document.body.style.overflow = 'hidden';
    });

    // Fermer le modal (croix ou clic sur overlay)
    closeModal?.addEventListener('click', closeModalHandler);
    window.addEventListener('click', (e) => {
        if (e.target === modal) {
            closeModalHandler();
        }
    });

    // Masquer le modal et réinitialiser l'état
    function closeModalHandler() {
        modal.classList.remove('active');
        document.body.style.overflow = 'auto';
        // Reset form
        resetModal();
    }

    // Sélection d'un plan d'abonnement (boutons "Choisir")
    paymentButtons.forEach(button => {
        button.addEventListener('click', () => {
            selectedPlan = button.getAttribute('data-plan');
            selectedPrice = button.getAttribute('data-price');
            selectedPlanName = button.closest('.plan').querySelector('h3').textContent;
            
            // Afficher le formulaire de paiement
            subscriptionPlans.style.display = 'none';
            paymentForm.classList.remove('hidden');
            
            // Mettre à jour l'affichage du prix formaté en FR
            const priceFormatted = parseFloat(selectedPrice).toLocaleString('fr-FR', {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2
            });
            totalPriceEl.textContent = priceFormatted + ' €';
            planNameEl.textContent = 'Plan: ' + selectedPlanName;
        });
    });

    // Retour à la grille des plans (bouton retour)
    backToPlans?.addEventListener('click', () => {
        subscriptionPlans.style.display = 'grid';
        paymentForm.classList.add('hidden');
        cardForm.reset();
    });

    // Soumission du formulaire de carte (simulé)
    cardForm?.addEventListener('submit', (e) => {
        e.preventDefault();
        
        // Récupérer les valeurs du formulaire
        const cardholderName = document.getElementById('cardholderName').value;
        const cardNumber = document.getElementById('cardNumber').value;
        const expiryDate = document.getElementById('expiryDate').value;
        const cvv = document.getElementById('cvv').value;
        const email = document.getElementById('email').value;

        // Validation basique (format uniquement)
        if (!validateCardNumber(cardNumber)) {
            showError('Numéro de carte invalide');
            return;
        }

        if (!validateExpiryDate(expiryDate)) {
            showError('Date d\'expiration invalide (MM/YY)');
            return;
        }

        if (!validateCVV(cvv)) {
            showError('CVV invalide');
            return;
        }

        // Simuler un traitement de paiement (API fictive)
        processPayment(cardholderName, cardNumber, expiryDate, cvv, email);
    });

    // Formatage du numéro de carte en groupes de 4 (XXXX XXXX XXXX XXXX)
    document.getElementById('cardNumber')?.addEventListener('input', (e) => {
        let value = e.target.value.replace(/\s/g, '');
        let formattedValue = '';
        for (let i = 0; i < value.length; i++) {
            if (i > 0 && i % 4 === 0) formattedValue += ' ';
            formattedValue += value[i];
        }
        e.target.value = formattedValue;
    });

    // Formatage de la date d'expiration MM/YY
    document.getElementById('expiryDate')?.addEventListener('input', (e) => {
        let value = e.target.value.replace(/\D/g, '');
        if (value.length >= 2) {
            value = value.slice(0, 2) + '/' + value.slice(2, 4);
        }
        e.target.value = value;
    });

    // CVV: n'accepter que des chiffres
    document.getElementById('cvv')?.addEventListener('input', (e) => {
        e.target.value = e.target.value.replace(/\D/g, '');
    });

    // Fonctions de validation (format uniquement; pas de Luhn réel ici)
    function validateCardNumber(cardNumber) {
        // Simple validation - just check if it has 16 digits
        const cleanNumber = cardNumber.replace(/\s/g, '');
        return /^\d{16}$/.test(cleanNumber);
    }

    function validateExpiryDate(date) {
        return /^\d{2}\/\d{2}$/.test(date);
    }

    function validateCVV(cvv) {
        return /^\d{3,4}$/.test(cvv);
    }

    // Traitement de paiement simulé et persistance localStorage
    function processPayment(name, cardNumber, expiry, cvv, email) {
        // Show loading state
        const submitBtn = cardForm.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;
        submitBtn.textContent = 'Traitement...';
        submitBtn.disabled = true;

        // Simulation d'appel API (2 secondes)
        setTimeout(() => {
            // Masquer le formulaire de paiement
            paymentForm.classList.add('hidden');
            
            // Afficher le message de succès
            successMessage.classList.remove('hidden');
            
            // Log d'information (dans une vraie app, on enverrait au serveur)
            console.log('Paiement réussi:', {
                plan: selectedPlan,
                planName: selectedPlanName,
                amount: selectedPrice,
                cardholderName: name,
                email: email,
                timestamp: new Date().toISOString()
            });

            // Stocker l'abonnement en localStorage (démo)
            const subscription = {
                plan: selectedPlan,
                planName: selectedPlanName,
                amount: selectedPrice,
                email: email,
                subscribedAt: new Date().toISOString(),
                status: 'active'
            };
            localStorage.setItem('groupie_subscription', JSON.stringify(subscription));

            // Reset button
            submitBtn.textContent = originalText;
            submitBtn.disabled = false;
        }, 2000);
    }

    // Fermer le message de succès
    closeSuccess?.addEventListener('click', closeModalHandler);

    // Affichage d'erreur simple (alert)
    function showError(message) {
        alert(message);
    }

    // Réinitialiser l'état du modal et du formulaire
    function resetModal() {
        subscriptionPlans.style.display = 'grid';
        paymentForm.classList.add('hidden');
        successMessage.classList.add('hidden');
        cardForm.reset();
        selectedPlan = null;
        selectedPrice = null;
        selectedPlanName = null;
    }

    // Vérifier l'abonnement existant au chargement
    function checkSubscriptionStatus() {
        const subscription = localStorage.getItem('groupie_subscription');
        if (subscription) {
            const data = JSON.parse(subscription);
            console.log('Utilisateur abonné:', data);
            // You can update UI based on subscription status here
        }
    }

    // Initialiser la vérification d'abonnement à DOMContentLoaded
    document.addEventListener('DOMContentLoaded', checkSubscriptionStatus);
})();
