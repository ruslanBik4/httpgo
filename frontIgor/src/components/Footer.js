import React from 'react'

const Footer = () => {
   return (
      <footer className="footer">
         <nav className="footer__nav">
            <div className="footer__nav-links">
               <ul>
                  <li>
                     <a href="/">Главная</a>
                  </li>
                  <li>
                     <a href="/">Отели</a>
                  </li>
                  <li>
                     <a href="/">Транспорт</a>
                  </li>
                  <li>
                     <a href="/">Отдых</a>
                  </li>
                  <li>
                     <a href="/">Путеводитель</a>
                  </li>
                  <li>
                     <a href="/">Блог</a>
                  </li>
                  <li>
                     <a href="/">Партнерам</a>
                  </li>
               </ul>
            </div>
            <div className="footer__contacts">
               <p>8-000-00-00 - free on country</p>
               <p>
                  +380(00)000-00-00 - according to the tariffs of your operator
               </p>
               <a href="mailto:webmaster@example.com">test@email.net</a>
               <br />
               <br />
               <ul>
                  <li>
                     <a href="/">F.A.Q.</a>
                  </li>
                  <li>
                     <a href="/">О нас</a>
                  </li>
               </ul>
            </div>
         </nav>
         <h3>Secured Systems</h3>
      </footer>
   );
}

export default Footer
