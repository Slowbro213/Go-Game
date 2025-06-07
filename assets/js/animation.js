export function createAnimator(character) {
  let targetX = 0;
  let targetY = 0;
  let currentX = 0;
  let currentY = 0;
  let animationId = null;
  let ID = null;

  function updatePosition(newX, newY) {
    targetX = newX;
    targetY = newY;
    if (!animationId) {
      animationId = requestAnimationFrame(animate);
    }
  }

  function animate() {
    const dx = targetX - currentX;
    const dy = targetY - currentY;
    
    if (Math.abs(dx) < 0.1 && Math.abs(dy) < 0.1) {
      currentX = targetX;
      currentY = targetY;
      animationId = null;
    } else {
      // Exponential smoothing
      currentX += dx * 0.2;
      currentY += dy * 0.2;
      animationId = requestAnimationFrame(animate);
    }

    character.style.transform = `translate(${currentX}px, ${currentY}px)`;
  }

  return { updatePosition };
}
