import { Component, OnDestroy, OnInit } from '@angular/core';
import { Accommodation } from '../model/accommodation';
import { AccommodationService } from '../service/accommodation.service';
import { Router } from '@angular/router';
import { AuthService } from '../service/auth.service';
import { Subscription, forkJoin, of } from 'rxjs';
import { FormControl, FormGroup } from '@angular/forms';
import { NgbDate } from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-main-page',
  templateUrl: './main-page.component.html',
  styleUrls: ['./main-page.component.css'],
})
export class MainPageComponent implements OnInit, OnDestroy {
  accommodations: Accommodation[] = [];
  isLoggedin: boolean = false;
  userRole: string = '';
  logSub: Subscription;
  rolesub: Subscription;

  searchAccoForm: FormGroup;
  locationInput: string = '';
  noGuestsInput: string = '';
  startDateInput: NgbDate | null = null
  endDateInout: NgbDate | null = null
  accomodationsByLocation: Accommodation[] = [];
  accomodationsByNoGuests: Accommodation[] = [];

  constructor(
    private router: Router,
    private accommodationService: AccommodationService,
    private authService: AuthService
  ) {
    this.logSub = this.authService.isLoggedin.subscribe(
      (data) => (this.isLoggedin = data)
    );
    this.rolesub = this.authService.role.subscribe(
      (data) => (this.userRole = data)
    );

    this.searchAccoForm = new FormGroup({
      location: new FormControl,
      startDate: new FormControl,
      endDate: new FormControl,
      noGuests: new FormControl
    })
  }

  ngOnInit(): void {
    this.authService.checkLoggin();
    this.authService.checkRole();

    this.accommodationService.getAll().subscribe({
      next: (data) => {
        this.accommodations = data as Accommodation[];
        // console.log(this.accommodations);
      },
      error: (err) => {
        console.log(err);
      },
    });
  }

  ngOnDestroy(): void {
    this.logSub.unsubscribe();
    this.rolesub.unsubscribe();
  }

  
searchAcco(): void {
  this.locationInput = this.searchAccoForm.get('location')?.value;
  this.startDateInput = this.searchAccoForm.get('startDate')?.value
  this.endDateInout = this.searchAccoForm.get('endDate')?.value
  this.noGuestsInput = this.searchAccoForm.get('noGuests')?.value;

  console.log(this.startDateInput)
  console.log(this.endDateInout)

  const locationObservable = this.locationInput != null && this.locationInput !== ''
    ? this.accommodationService.getAllByLocation(this.locationInput)
    : of([]);

  const noGuestsObservable = this.noGuestsInput != null && this.noGuestsInput !== ''
    ? this.accommodationService.getAllByNoGuests(this.noGuestsInput)
    : of([]);

  forkJoin([locationObservable, noGuestsObservable]).subscribe({
    next: ([locations, noGuests]: [Accommodation[], Accommodation[]]) => {
      this.accomodationsByLocation = locations as Accommodation[];
      this.accomodationsByNoGuests = noGuests as Accommodation[];

      if (this.accomodationsByLocation.length > 0 && this.accomodationsByNoGuests.length == 0) {
        console.log("Ima lokacija nema gostiju");
        this.accommodations = this.accomodationsByLocation;
      } 
      else if (this.accomodationsByLocation.length == 0 && this.accomodationsByNoGuests.length > 0) {
        console.log("Nema lokacija ima gostiju");
        this.accommodations = this.accomodationsByNoGuests;
      } 
      else if (this.accomodationsByLocation.length > 0 && this.accomodationsByNoGuests.length > 0) {
        console.log("Ima oba");
        const tempList: Accommodation[] = [];
        for (const accoLocation of this.accomodationsByLocation){
          for (const accoNoGuest of this.accomodationsByNoGuests){
            if (accoLocation._id == accoNoGuest._id){
              tempList.push(accoLocation);
            }
            else{
              continue;
            }
          }
        }
        this.accommodations = tempList
      } 
      else if (this.accomodationsByLocation.length == 0 && this.accomodationsByNoGuests.length == 0) {
        this.ngOnInit();
      }

      this.accomodationsByLocation = [];
      this.accomodationsByNoGuests = [];
    },
    error: (err) => {
      console.log(err);
    },
    complete: () => {
      console.log('Both observables complete');
    },
  });
}
  //Used for filtering all searched results
  // findCommonElements<T>(arrays: T[][]): T[] {
  //   if (arrays.length === 0) {
  //     return [];
  //   }
  
  //   // Use the first array as the base for comparison
  //   const baseArray = arrays[0];
  
  //   // Filter elements that are present in all arrays
  //   const commonElements = baseArray.filter((element) =>
  //     arrays.every((array) => array.includes(element))
  //   );
  
  //   return commonElements;
  // }
}
