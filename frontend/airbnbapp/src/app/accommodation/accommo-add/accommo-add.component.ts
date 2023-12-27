import { Component } from '@angular/core';
import {
  AbstractControl,
  FormArray,
  FormBuilder,
  FormControl,
  FormGroup,
  Validators,
} from '@angular/forms';
import { Router } from '@angular/router';
import { AccommodationService } from 'src/app/service/accommodation.service';
import { Accommodation } from '../../model/accommodation';
import { AmenityType } from '../../model/amenityType';
import { AuthService } from '../../service/auth.service';

@Component({
  selector: 'app-accommo-add',
  templateUrl: './accommo-add.component.html',
  styleUrls: ['./accommo-add.component.css'],
})
export class AccommoAddComponent {
  form: FormGroup;
  amenityRange = AmenityType;

  constructor(
    private fb: FormBuilder,
    private router: Router,
    private accommodationService: AccommodationService,
    private authService: AuthService
  ) {
    this.form = this.fb.group({
      name: [null, Validators.required],
      location: this.fb.group({
        country: [null, Validators.required],
        city: [null, Validators.required],
        streetName: [null, Validators.required],
        streetNumber: [null, Validators.required],
      }),
      minGuests: [null, Validators.required],
      maxGuests: [null, Validators.required],
      amenities: new FormArray([], Validators.required),
      // price: [null, Validators.required],
    });
  }

  submit() {
    // const accommodation = <Accommodation>{ ...this.form.value };
    let accommodation: Accommodation = this.form.value;
    accommodation.username = this.authService.getUsername();
    // console.log(accommodation);
    this.accommodationService.insert(accommodation).subscribe({
      next: (data) => {
        console.log('create success');
        this.router.navigate(['']);
      },
      error: (err) => {
        console.log(err);
      },
    });
  }

  getRange(obj: any) {
    return Object.values(obj);
  }

  onCheckChange(event: any) {
    const formArray: FormArray = this.form.get('amenities') as FormArray;

    if (event.target.checked) {
      formArray.push(new FormControl(event.target.value));
    } else {
      let i: number = 0;

      formArray.controls.forEach((ctrl: AbstractControl<any>) => {
        if (ctrl.value == event.target.value) {
          formArray.removeAt(i);
          return;
        }
        i++;
      });
    }
  }
}
