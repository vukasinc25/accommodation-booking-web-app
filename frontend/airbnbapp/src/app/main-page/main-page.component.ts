import { Component, OnDestroy, OnInit } from '@angular/core';
import { Accommodation } from '../model/accommodation';
import { AccommodationService } from '../service/accommodation.service';
import { Router } from '@angular/router';
import { AuthService } from '../service/auth.service';
import { Subscription, catchError, forkJoin, of } from 'rxjs';
import { AbstractControl, FormArray, FormControl, FormGroup } from '@angular/forms';
import { NgbDate } from '@ng-bootstrap/ng-bootstrap';
import { ReservationService } from '../service/reservation.service';
import { ReservationByDateSearch } from '../model/reservationByDateSearch';
import { AmenityType } from '../model/amenityType';

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
  
  //Search
  searchAccoForm: FormGroup;
  filterAccoForm: FormGroup;
  locationInput: string = '';
  noGuestsInput: string = '';
  startDateInput: string = '';
  endDateInput: string = '';
  accommodationsByLocation: Accommodation[] = [];
  accommodationsByNoGuest: Accommodation[] = [];
  accommodationsByDate: Accommodation[] = [];

  //Filter
  amenityRange = AmenityType;
  priceFrom: number = 0;
  priceTo: number = 0;
  isFeatured: boolean = false;
  amenities: [] = [];
  userId: 0;

  constructor(
    private router: Router,
    private accommodationService: AccommodationService,
    private authService: AuthService,
    private reservationService: ReservationService
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

    this.filterAccoForm = new FormGroup({
      priceFrom: new FormControl,
      priceTo: new FormControl,
      amenities: new FormArray([]),
      isFeatured: new FormControl
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

  filterAcco(): void {
    this.priceFrom = this.filterAccoForm.get('priceFrom')?.value;
    this.priceTo = this.filterAccoForm.get('priceTo')?.value;
    this.amenities = this.filterAccoForm.get('amenities')?.value
    this.isFeatured = this.filterAccoForm.get('isFeatured')?.value;

    if (this.priceFrom > 0 || this.priceTo > 0) {
      for (const accommodation of this.accommodations){
        this.authService.(accommodation.username)
      }
      this.reservationService.getReservationsByAccoId()
    }
    
    console.log(this.filterAccoForm.value)
  }

  searchAcco(): void {
    this.locationInput = this.searchAccoForm.get('location')?.value;
    this.startDateInput = this.searchAccoForm.get('startDate')?.value
    this.endDateInput = this.searchAccoForm.get('endDate')?.value
    this.noGuestsInput = this.searchAccoForm.get('noGuests')?.value;

    const locationObservable = this.locationInput != null && this.locationInput !== ''
      ? this.accommodationService.getAllByLocation(this.locationInput)
      : of([]);

    const noGuestsObservable = this.noGuestsInput != null && this.noGuestsInput !== ''
      ? this.accommodationService.getAllByNoGuests(this.noGuestsInput)
      : of([]);

    const dateObservable = this.startDateInput != null && this.endDateInput != null
      ? this.accommodationService.getAllByDate(this.startDateInput, this.endDateInput)
      : of([]);

    forkJoin([locationObservable, noGuestsObservable, dateObservable])
    // .pipe(
    //   catchError(error => {
    //     console.error('Error occurred in one of the observables:', error);
    //     return of([[], [], []]); // Return default values for all observables
    //   })
    // )
    .subscribe({
      next: ([locations, noGuests, accoDate]: [Accommodation[], Accommodation[], ReservationByDateSearch[]]) => {
        this.accommodationsByLocation = locations as Accommodation[];
        this.accommodationsByNoGuest = noGuests as Accommodation[];
        this.accommodationsByDate = accoDate as Accommodation[];
        console.log(this.accommodationsByDate)

        if (this.accommodationsByLocation.length > 0 && this.accommodationsByNoGuest.length == 0 && this.accommodationsByDate.length == 0) {
          console.log("Search by location only");
          this.accommodations = this.accommodationsByLocation;
        }
        else if (this.accommodationsByLocation.length > 0 && this.accommodationsByDate.length > 0 && this.accommodationsByNoGuest.length > 0) {
          console.log("Search by all three");
          const tempList: Accommodation[] = [];
          for (const accoNoGuest of this.accommodationsByNoGuest){
            for (const accoDate of this.accommodationsByDate){
              for (const accoLocation of this.accommodationsByLocation){
                if (accoLocation._id == accoNoGuest._id && accoLocation._id == accoDate._id){
                  tempList.push(accoLocation);
                }
                else{
                  continue;
                }
              }
            }
          }
          this.accommodations = tempList
        }
        else if (this.accommodationsByDate.length > 0 && this.accommodationsByNoGuest.length == 0 && this.accommodationsByLocation.length == 0) {
          console.log("Search by date only");
          this.accommodations = this.accommodationsByDate;
        } 
        else if (this.accommodationsByLocation.length == 0 && this.accommodationsByNoGuest.length > 0 && this.accommodationsByDate.length == 0) {
          console.log("Search by number of guests only");
          this.accommodations = this.accommodationsByNoGuest;
        } 
        else if (this.accommodationsByLocation.length > 0 && this.accommodationsByNoGuest.length > 0) {
          console.log("Search by location and number of guests");
          const tempList: Accommodation[] = [];
          for (const accoLocation of this.accommodationsByLocation){
            for (const accoNoGuest of this.accommodationsByNoGuest){
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
        else if (this.accommodationsByLocation.length > 0 && this.accommodationsByDate.length > 0 && this.accommodationsByNoGuest.length == 0) {
          console.log("Search by location and date");
          const tempList: Accommodation[] = [];
          for (const accoLocation of this.accommodationsByLocation){
            for (const accoNoGuest of this.accommodationsByDate){
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
        else if (this.accommodationsByLocation.length == 0 && this.accommodationsByDate.length > 0 && this.accommodationsByNoGuest.length > 0) {
          console.log("Search by date and number of guests");
          const tempList: Accommodation[] = [];
          for (const accoLocation of this.accommodationsByNoGuest){
            for (const accoNoGuest of this.accommodationsByDate){
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
        else if (this.accommodationsByLocation.length == 0 && this.accommodationsByNoGuest.length == 0 && this.accommodationsByDate.length == 0) {
          this.ngOnInit();
        }

        this.accommodationsByLocation = [];
        this.accommodationsByNoGuest = [];
        this.accommodationsByDate = [];
      },
      error: (err) => {
        console.log(err);
      },
      complete: () => {
        console.log('Both observables complete');
      },
    });
  }

  getRange(obj: any) {
    return Object.values(obj);
  }

  onCheckChange(event: any) {
    const formArray: FormArray = this.filterAccoForm.get('amenities') as FormArray;

    if (event.target.checked) {
      formArray.push(new FormControl(event.target.value));
    } else {
      let i: number = 0;

      formArray.controls.forEach(
        (ctrl: AbstractControl<any>, index: number) => {
          if (ctrl.value == event.target.value) {
            formArray.removeAt(index);
            return;
          }
        }
      );
    }
  }
}
