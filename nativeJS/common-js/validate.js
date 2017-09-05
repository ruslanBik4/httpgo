
const classNameError = 'error-validate';
const attr = 'data-validate';

const attrSelect = 'data-validate-select'; // data-validate-select="first, 1, 2, ..., last" => value attr to option
const attrCheckbox = 'data-validate-checkbox';

export class Validate {

  static validate(data, isTest = false) {
    let isError = false;

    data.querySelectorAll(`[${ attr }]`).forEach((component) => {

      let componentChild = component.querySelector('select');

      // select

      if (componentChild) {
        const validateSelect = component.getAttribute(attrSelect);

        if (validateSelect) {

          let result = false;
          const indexSelected = componentChild.selectedIndex;
          const valueSelect = (indexSelected > -1) ? componentChild.options[indexSelected].value : false;

          if (valueSelect) {
            validateSelect.split(', ').forEach((value) => {
              if (value === 'first') {
                result = (indexSelected == 0);
              } else if (value === 'last') {
                result = (indexSelected == componentChild.options.length - 1);
              } else {
                result = (value == valueSelect);
              }
              if (result) {
                return false;
              }
            });
          }

          if (result) {
            if (!isTest) componentChild.classList.add(classNameError);
            isError = true;
          } else {
            componentChild.classList.remove(classNameError);
          }

        }
        return;
      }

      // input

      if (component.hasAttribute(attrCheckbox)) {
        let result = false;
        component.querySelectorAll('input[type="checkbox"]').forEach((input) => {
          if (input.checked) {
            result = true;
          }
        });
        isError = !(result);
        return;
      }

      componentChild = component.querySelector('input');

      if (componentChild) {
        if (componentChild.value.length === 0) {
          if (!isTest) componentChild.classList.add(classNameError);
          isError = true;
        } else {
          componentChild.classList.remove(classNameError);
        }
        return;
      }

      //textarea

      componentChild = component.querySelector('textarea');

      if (componentChild) {
        if (componentChild.textLength === 0) {
          if (!isTest) componentChild.classList.add(classNameError);
          isError = true;
        } else {
          componentChild.classList.remove(classNameError);
        }
      }

    });

    return isError;
  }

}