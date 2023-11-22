
let scrollToTopBtn = document.getElementById("btn-scroll-to-top");
let fourViewports = window.innerHeight * 4;

window.onscroll = function () {
  scrollFunction();
};

function scrollFunction() {
  if (
    document.body.scrollTop > fourViewports ||
    document.documentElement.scrollTop > fourViewports
  ) {
    scrollToTopBtn.style.display = "block";
  } else {
    scrollToTopBtn.style.display = "none";
  }
}

scrollToTopBtn.addEventListener("click", backToTop);

function backToTop() {
  document.body.scrollTop = 0;
  document.documentElement.scrollTop = 0;
}
