import React from "react";
import Logo from "../images/Logo.jpeg";

const Header = () => {
   return (
      <header className="header">
         <div className="header__container">
            <a href="/">
               <img src={Logo} alt="logo" className="header__logo-img" />
            </a>
            <nav className="header__nav">
               <a href="/">Autorization</a>
               <a href="/">View</a>
               <div>
                  <label htmlFor="search">Search</label>
                  <input type="text" name="search" id="search" />
               </div>
            </nav>
         </div>
      </header>
   );
};

export default Header;
